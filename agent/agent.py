import asyncio
import json
import os
import platform
import socket
import uuid
from dataclasses import dataclass
from fractions import Fraction
from pathlib import Path
from typing import Any, Dict

import aiohttp
import mss
import numpy as np
from aiortc import RTCConfiguration, RTCIceServer, RTCPeerConnection, RTCSessionDescription, VideoStreamTrack
from aiortc.sdp import candidate_from_sdp
from av.video.frame import VideoFrame
from pynput import keyboard, mouse


@dataclass
class AgentConfig:
    server_http: str
    server_ws: str
    username: str
    password: str
    device_id: str
    device_name: str
    heartbeat_sec: int = 10


class ScreenTrack(VideoStreamTrack):
    """Capture the selected monitor and expose it as a WebRTC video track."""

    def __init__(self, display_id: int, fps: int, quality: str) -> None:
        super().__init__()
        self.sct = mss.mss()
        self.display_id = max(1, int(display_id))
        self.monitor = self._pick_monitor(self.display_id)
        self.fps = max(5, min(int(fps), 60))
        self.frame_interval = 1.0 / self.fps
        self.quality = quality
        self.scale = self._quality_scale(quality)
        self._next_frame_at = 0.0

    def _pick_monitor(self, display_id: int) -> Dict[str, int]:
        monitors = self.sct.monitors
        if display_id >= len(monitors):
            display_id = 1
        return monitors[display_id]

    @staticmethod
    def _quality_scale(quality: str) -> float:
        if quality == "low":
            return 0.5
        if quality == "medium":
            return 0.75
        return 1.0

    def update_stream_params(self, fps: int, quality: str) -> None:
        self.fps = max(5, min(int(fps), 60))
        self.frame_interval = 1.0 / self.fps
        self.quality = quality if quality in {"low", "medium", "high", "ultra"} else "medium"
        self.scale = self._quality_scale(self.quality)

    async def recv(self) -> VideoFrame:
        pts, time_base = await self.next_timestamp()

        now = asyncio.get_running_loop().time()
        if self._next_frame_at > now:
            await asyncio.sleep(self._next_frame_at - now)
        self._next_frame_at = asyncio.get_running_loop().time() + self.frame_interval

        frame = np.asarray(self.sct.grab(self.monitor))[:, :, :3]
        video = VideoFrame.from_ndarray(frame, format="bgr24")

        if self.scale < 0.99:
            width = int(video.width * self.scale)
            height = int(video.height * self.scale)
            width = max(width, 160)
            height = max(height, 90)
            video = video.reformat(width=width, height=height)

        video.pts = pts
        video.time_base = time_base if time_base is not None else Fraction(1, 90000)
        return video

    def stop(self) -> None:
        super().stop()
        self.sct.close()


class ControlInjector:
    def __init__(self, monitor: Dict[str, int]) -> None:
        self.monitor = monitor
        self.mouse = mouse.Controller()
        self.keyboard = keyboard.Controller()

    def handle(self, event: Dict[str, Any]) -> None:
        kind = event.get("kind")
        if kind == "mouse_move":
            self._mouse_move(event)
        elif kind == "mouse_click":
            self._mouse_click(event)
        elif kind == "wheel":
            self._wheel(event)
        elif kind == "key":
            self._key(event)
        elif kind == "text":
            self._text(event)

    def _mouse_move(self, event: Dict[str, Any]) -> None:
        x = float(event.get("x", 0.0))
        y = float(event.get("y", 0.0))
        x = max(0.0, min(1.0, x))
        y = max(0.0, min(1.0, y))

        abs_x = int(self.monitor["left"] + x * self.monitor["width"])
        abs_y = int(self.monitor["top"] + y * self.monitor["height"])
        self.mouse.position = (abs_x, abs_y)

    def _mouse_click(self, event: Dict[str, Any]) -> None:
        button_name = event.get("button", "left")
        pressed = bool(event.get("pressed", True))
        btn = {
            "left": mouse.Button.left,
            "right": mouse.Button.right,
            "middle": mouse.Button.middle,
        }.get(button_name, mouse.Button.left)

        if pressed:
            self.mouse.press(btn)
        else:
            self.mouse.release(btn)

    def _wheel(self, event: Dict[str, Any]) -> None:
        dx = int(event.get("dx", 0))
        dy = int(event.get("dy", 0))
        if dx == 0 and dy == 0:
            return

        # Keep wheel deltas bounded to avoid accidental jumps.
        dx = max(-20, min(20, dx))
        dy = max(-20, min(20, dy))
        self.mouse.scroll(dx, dy)

    def _key(self, event: Dict[str, Any]) -> None:
        key_name = str(event.get("key", ""))
        if not key_name:
            return
        pressed = bool(event.get("pressed", True))

        key_obj = self._resolve_key(key_name)
        if key_obj is None:
            return

        if pressed:
            self.keyboard.press(key_obj)
        else:
            self.keyboard.release(key_obj)

    def _text(self, event: Dict[str, Any]) -> None:
        text = str(event.get("text", ""))
        if not text:
            return
        # Limit one payload size to avoid accidental large paste storms.
        self.keyboard.type(text[:256])

    @staticmethod
    def _resolve_key(key_name: str) -> Any:
        special = {
            "Enter": keyboard.Key.enter,
            "Backspace": keyboard.Key.backspace,
            "Tab": keyboard.Key.tab,
            "Escape": keyboard.Key.esc,
            "ArrowUp": keyboard.Key.up,
            "ArrowDown": keyboard.Key.down,
            "ArrowLeft": keyboard.Key.left,
            "ArrowRight": keyboard.Key.right,
            "Shift": keyboard.Key.shift,
            "Control": keyboard.Key.ctrl,
            "Alt": keyboard.Key.alt,
            "Meta": keyboard.Key.cmd,
            "Delete": keyboard.Key.delete,
            "Home": keyboard.Key.home,
            "End": keyboard.Key.end,
            "PageUp": keyboard.Key.page_up,
            "PageDown": keyboard.Key.page_down,
            " ": keyboard.Key.space,
        }
        if key_name in special:
            return special[key_name]
        if len(key_name) == 1:
            return key_name
        return None


class RemoteAgent:
    def __init__(self, config: AgentConfig) -> None:
        self.cfg = config
        self.token = ""
        self.ws: aiohttp.ClientWebSocketResponse | None = None
        self.pc_by_session: Dict[str, RTCPeerConnection] = {}
        self.track_by_session: Dict[str, ScreenTrack] = {}
        self.injector_by_session: Dict[str, ControlInjector] = {}

    async def run(self) -> None:
        while True:
            try:
                await self._connect_once()
            except Exception as exc:  # noqa: BLE001
                print(f"[agent] disconnected: {exc}")
            await self._cleanup_all_sessions()
            await asyncio.sleep(3)

    async def _connect_once(self) -> None:
        self.token = await self._login()

        timeout = aiohttp.ClientTimeout(total=30)
        async with aiohttp.ClientSession(timeout=timeout) as session:
            async with session.ws_connect(self.cfg.server_ws) as ws:
                self.ws = ws
                await ws.send_json({"type": "auth", "token": self.token})
                await self._register_device()
                heartbeat = asyncio.create_task(self._heartbeat_loop())

                try:
                    async for msg in ws:
                        if msg.type != aiohttp.WSMsgType.TEXT:
                            continue
                        payload = json.loads(msg.data)
                        await self._handle_message(payload)
                finally:
                    heartbeat.cancel()
                    self.ws = None

    async def _login(self) -> str:
        timeout = aiohttp.ClientTimeout(total=15)
        async with aiohttp.ClientSession(timeout=timeout) as session:
            async with session.post(
                f"{self.cfg.server_http}/api/login",
                json={"username": self.cfg.username, "password": self.cfg.password},
            ) as resp:
                data = await resp.json()
                if resp.status != 200:
                    raise RuntimeError(f"login failed: {data}")
                return data["token"]

    async def _register_device(self) -> None:
        assert self.ws is not None
        with mss.mss() as sct:
            displays = []
            for index, monitor in enumerate(sct.monitors[1:], start=1):
                displays.append(
                    {
                        "id": index,
                        "label": f"Display {index}",
                        "width": int(monitor["width"]),
                        "height": int(monitor["height"]),
                    }
                )

        await self.ws.send_json(
            {
                "type": "register_device",
                "device": {
                    "id": self.cfg.device_id,
                    "name": self.cfg.device_name,
                    "os": platform.platform(),
                    "displays": displays,
                },
            }
        )

    async def _heartbeat_loop(self) -> None:
        while True:
            await asyncio.sleep(self.cfg.heartbeat_sec)
            if self.ws is None:
                return
            await self.ws.send_json({"type": "heartbeat"})

    async def _handle_message(self, msg: Dict[str, Any]) -> None:
        msg_type = msg.get("type")
        if msg_type == "device_registered":
            self.cfg.device_id = msg.get("device_id", self.cfg.device_id)
            print(f"[agent] registered as {self.cfg.device_id}")
            return

        if msg_type == "session_offer_request":
            await self._start_session(msg)
            return

        if msg_type == "signal":
            await self._handle_signal(msg)
            return

        if msg_type == "session_update":
            await self._handle_session_update(msg)
            return

        if msg_type == "control":
            await self._handle_control(msg)
            return

        if msg_type == "session_ended":
            session_id = str(msg.get("session_id", ""))
            await self._close_session(session_id)
            return

        if msg_type == "error":
            print(f"[agent] server error: {msg.get('message')}")

    async def _start_session(self, msg: Dict[str, Any]) -> None:
        session_id = str(msg.get("session_id", ""))
        if not session_id:
            return

        await self._close_session(session_id)

        display_id = int(msg.get("display_id", 1))
        fps = int(msg.get("fps", 15))
        quality = str(msg.get("quality", "medium"))
        webrtc = msg.get("webrtc", {}) or {}
        transport_mode = str(webrtc.get("mode", "p2p"))
        print(f"[agent] session={session_id} transport={transport_mode}")

        pc = RTCPeerConnection(configuration=self._build_rtc_configuration(webrtc))
        track = ScreenTrack(display_id=display_id, fps=fps, quality=quality)
        injector = ControlInjector(track.monitor)

        self.pc_by_session[session_id] = pc
        self.track_by_session[session_id] = track
        self.injector_by_session[session_id] = injector

        pc.addTrack(track)

        @pc.on("icecandidate")
        async def on_icecandidate(candidate):
            if candidate is None:
                return
            candidate_sdp = candidate.to_sdp()
            if not candidate_sdp.startswith("candidate:"):
                candidate_sdp = "candidate:" + candidate_sdp
            await self._send_signal(
                session_id,
                {
                    "type": "candidate",
                    "candidate": {
                        "candidate": candidate_sdp,
                        "sdpMid": candidate.sdpMid,
                        "sdpMLineIndex": candidate.sdpMLineIndex,
                    },
                },
            )

        @pc.on("connectionstatechange")
        async def on_state_change():
            state = pc.connectionState
            print(f"[agent] session={session_id} state={state}")
            if state in {"failed", "closed", "disconnected"}:
                await self._close_session(session_id)

        offer = await pc.createOffer()
        bitrate_kbps = self._video_bitrate_kbps(quality)
        tuned_sdp = self._apply_video_bandwidth(offer.sdp, bitrate_kbps)
        await pc.setLocalDescription(RTCSessionDescription(sdp=tuned_sdp, type=offer.type))
        await self._send_signal(
            session_id,
            {
                "type": "offer",
                "sdp": pc.localDescription.sdp,
            },
        )

    async def _handle_signal(self, msg: Dict[str, Any]) -> None:
        session_id = str(msg.get("session_id", ""))
        payload = msg.get("payload", {})
        if not session_id or session_id not in self.pc_by_session:
            return

        pc = self.pc_by_session[session_id]
        signal_type = payload.get("type")

        if signal_type == "answer":
            sdp = payload.get("sdp")
            if not sdp:
                return
            await pc.setRemoteDescription(RTCSessionDescription(sdp=sdp, type="answer"))
            return

        if signal_type == "candidate":
            candidate = payload.get("candidate") or {}
            cstr = candidate.get("candidate")
            if not cstr:
                return
            if cstr.startswith("candidate:"):
                cstr = cstr.split(":", 1)[1]
            rtc_candidate = candidate_from_sdp(cstr)
            rtc_candidate.sdpMid = candidate.get("sdpMid")
            rtc_candidate.sdpMLineIndex = candidate.get("sdpMLineIndex")
            await pc.addIceCandidate(rtc_candidate)

    async def _handle_session_update(self, msg: Dict[str, Any]) -> None:
        session_id = str(msg.get("session_id", ""))
        if not session_id:
            return

        track = self.track_by_session.get(session_id)
        if track is None:
            return

        fps = int(msg.get("fps", track.fps))
        quality = str(msg.get("quality", track.quality)).lower()
        track.update_stream_params(fps=fps, quality=quality)
        print(f"[agent] session={session_id} updated fps={track.fps} quality={track.quality}")

    async def _handle_control(self, msg: Dict[str, Any]) -> None:
        session_id = str(msg.get("session_id", ""))
        event = msg.get("event", {})
        if not session_id:
            return

        if str(event.get("kind", "")) == "stream_config":
            track = self.track_by_session.get(session_id)
            if track is None:
                return
            fps = int(event.get("fps", track.fps))
            quality = str(event.get("quality", track.quality)).lower()
            track.update_stream_params(fps=fps, quality=quality)
            print(f"[agent] session={session_id} stream_config fps={track.fps} quality={track.quality}")
            return

        injector = self.injector_by_session.get(session_id)
        if not injector:
            return

        try:
            injector.handle(event)
        except Exception as exc:  # noqa: BLE001
            print(f"[agent] control event failed: {exc}")

    async def _send_signal(self, session_id: str, payload: Dict[str, Any]) -> None:
        if self.ws is None:
            return
        await self.ws.send_json(
            {
                "type": "signal",
                "session_id": session_id,
                "payload": payload,
            }
        )

    async def _close_session(self, session_id: str) -> None:
        if not session_id:
            return

        pc = self.pc_by_session.pop(session_id, None)
        track = self.track_by_session.pop(session_id, None)
        self.injector_by_session.pop(session_id, None)

        if track is not None:
            track.stop()
        if pc is not None:
            await pc.close()

    async def _cleanup_all_sessions(self) -> None:
        for session_id in list(self.pc_by_session.keys()):
            await self._close_session(session_id)

    @staticmethod
    def _build_rtc_configuration(webrtc: Dict[str, Any]) -> RTCConfiguration:
        ice_servers: list[RTCIceServer] = []
        for item in webrtc.get("ice_servers", []) or []:
            urls = item.get("urls", [])
            if isinstance(urls, str):
                urls = [urls]
            if not isinstance(urls, list) or len(urls) == 0:
                continue

            username = item.get("username")
            credential = item.get("credential")
            kwargs: Dict[str, Any] = {"urls": urls}
            if username:
                kwargs["username"] = str(username)
            if credential:
                kwargs["credential"] = str(credential)
            ice_servers.append(RTCIceServer(**kwargs))

        return RTCConfiguration(iceServers=ice_servers)

    @staticmethod
    def _video_bitrate_kbps(quality: str) -> int:
        q = str(quality).lower()
        if q == "low":
            return 1200
        if q == "medium":
            return 2500
        if q == "high":
            return 5000
        return 9000

    @staticmethod
    def _apply_video_bandwidth(sdp: str, bitrate_kbps: int) -> str:
        lines = sdp.splitlines()
        out: list[str] = []
        in_video = False

        for line in lines:
            if line.startswith("m="):
                in_video = line.startswith("m=video")
                out.append(line)
                if in_video:
                    out.append(f"b=AS:{max(300, int(bitrate_kbps))}")
                continue

            if in_video and line.startswith("b=AS:"):
                continue

            out.append(line)

        return "\r\n".join(out) + "\r\n"


def load_config() -> AgentConfig:
    load_env_file(os.getenv("ENV_FILE", ".env"))

    server_http = os.getenv("SERVER_HTTP", "http://127.0.0.1:8080").rstrip("/")
    default_ws = server_http.replace("http://", "ws://").replace("https://", "wss://") + "/ws"
    server_ws = os.getenv("SERVER_WS", default_ws)

    username = os.getenv("AGENT_USER", "agent")
    password = os.getenv("AGENT_PASS", "agent123")

    device_id = os.getenv("DEVICE_ID", f"{socket.gethostname()}-{uuid.uuid4().hex[:8]}")
    device_name = os.getenv("DEVICE_NAME", platform.node() or "PythonAgent")

    return AgentConfig(
        server_http=server_http,
        server_ws=server_ws,
        username=username,
        password=password,
        device_id=device_id,
        device_name=device_name,
    )


def load_env_file(path: str) -> None:
    file_path = Path(path.strip() or ".env")
    if not file_path.exists():
        return

    for line_no, raw in enumerate(file_path.read_text(encoding="utf-8").splitlines(), start=1):
        line = raw.strip()
        if not line or line.startswith("#"):
            continue
        if line.startswith("export "):
            line = line[len("export ") :].strip()

        if "=" not in line:
            raise ValueError(f"invalid env line {line_no} in {file_path}")

        key, value = line.split("=", 1)
        key = key.strip()
        value = value.strip()
        if not key:
            raise ValueError(f"empty env key at line {line_no} in {file_path}")

        if (value.startswith("\"") and value.endswith("\"")) or (
            value.startswith("'") and value.endswith("'")
        ):
            value = value[1:-1]

        # Existing process env has higher priority than file values.
        os.environ.setdefault(key, value)


async def main() -> None:
    cfg = load_config()
    print(f"[agent] connect server={cfg.server_ws} device={cfg.device_name}")
    agent = RemoteAgent(cfg)
    await agent.run()


if __name__ == "__main__":
    asyncio.run(main())
