import { useState, useEffect, useCallback, useRef } from 'react';

// Define a complete and safe initial state for the application.
// This ensures no component receives 'undefined' on the first render.
const initialStatus = {
  state: 'connecting', // A valid string state from the start
  isAutoRec: false,
  diskUsage: 0,
  cpuUsage: 0,
  memoryUsage: 0,
};

const initialConfig = {
  web: {},
  audio: {},
  autoRecord: { enabled: false },
  icecast: { enabled: false },
  srt: { enabled: false },
};

export const useNixonApi = () => {
  const [isConnected, setIsConnected] = useState(false);
  const [appStatus, setAppStatus] = useState(initialStatus);
  const [config, setConfig] = useState(initialConfig);
  const [recordings, setRecordings] = useState([]);
  const [audioCaps, setAudioCaps] = useState(null);

  const socketRef = useRef(null); // This will hold the active WebSocket instance

  useEffect(() => {
    const token = import.meta.env.VITE_WS_SECRET || "nixon-default-secret";
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const socketUrl = `${wsProtocol}//${window.location.host}/ws?token=${token}`;

    console.log('Attempting WebSocket connection...');
    const ws = new WebSocket(socketUrl); // Create new WebSocket instance for this effect run

    ws.onopen = () => {
      setIsConnected(true);
      console.log('WebSocket connection opened');
    };

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        if (data.status) {
          setAppStatus(data.status);
        }
        if (data.config) {
          setConfig(data.config);
        }
        if (data.recordings) {
          setRecordings(data.recordings);
        }
        if (data.audioCaps) {
            setAudioCaps(data.audioCaps);
        }
      } catch (e) {
        console.error("Failed to parse WebSocket message:", e, event.data);
      }
    };

    ws.onclose = (event) => {
      setIsConnected(false);
      console.log('WebSocket connection closed', event);
      // Removed auto-reconnect here. If a reconnect is desired, it should be managed
      // by a parent component's state or a more sophisticated custom hook's retry logic
      // that causes this useEffect to re-run, rather than directly within onclose.
    };

    ws.onerror = (error) => {
      console.error("WebSocket error:", error);
      // It's crucial to explicitly close on error to ensure onclose handler is called
      if (ws.readyState !== WebSocket.CLOSED) {
        ws.close();
      }
    };

    socketRef.current = ws; // Update the ref to the *current* active WebSocket instance

    // Cleanup function: close the WebSocket when component unmounts or effect re-runs
    return () => {
      console.log('WebSocket cleanup: Closing connection for effect cleanup');
      if (ws.readyState !== WebSocket.CLOSED) { // Close the specific 'ws' instance created by this effect run
        ws.close(1000, "Component unmounted/effect cleanup"); // Code 1000 for normal closure
      }
      socketRef.current = null; // Clear the ref on cleanup
    };
  }, []); // Empty dependency array: effect runs once on mount, cleans up on unmount

  const sendCommand = useCallback((command, payload) => {
    if (socketRef.current && socketRef.current.readyState === WebSocket.OPEN) {
      socketRef.current.send(JSON.stringify({ command, payload }));
    } else {
      console.warn("WebSocket is not open. Command not sent:", command);
    }
  }, []); // No dependencies needed for sendCommand as it uses socketRef.current

  return { isConnected, appStatus, config, recordings, audioCaps, sendCommand };
};
