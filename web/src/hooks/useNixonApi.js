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

  const socketRef = useRef(null);

  const connectWebSocket = useCallback(() => {
    const token = "nixon-default-secret";
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'; // ADDED/MODIFIED: Determine secure/insecure protocol
    const socketUrl = `${wsProtocol}//${window.location.host}/ws?token=${token}`; // MODIFIED: Use wsProtocol
        if (socketRef.current && socketRef.current.readyState < 2) return;

    socketRef.current = new WebSocket(socketUrl);
    socketRef.current.onopen = () => setIsConnected(true);
    socketRef.current.onclose = () => {
      setIsConnected(false);
      setTimeout(connectWebSocket, 3000);
    };
    socketRef.current.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        if (message.type === 'status_update') {
          setAppStatus(prev => ({ ...prev, ...message.payload }));
        }
      } catch (error) {
        console.error("Error parsing WebSocket message:", error);
      }
    };
    socketRef.current.onerror = (error) => console.error("WebSocket error:", error);
  }, []);

  useEffect(() => {
    fetch('/api/status').then(res => res.json()).then(data => setAppStatus(data || initialStatus));
    fetch('/api/recordings').then(res => res.json()).then(data => setRecordings(data || []));
    
    connectWebSocket();

    return () => {
      if (socketRef.current) {
        socketRef.current.onclose = null;
        socketRef.current.close();
      }
    };
  }, [connectWebSocket]);

  const toggleRecording = () => {
    const endpoint = appStatus?.state === 'recording' ? '/api/recording/stop' : '/api/recording/start';
    fetch(endpoint, { method: 'POST' });
  };

  return {
    appStatus, config, recordings, audioCaps, isConnected,
    handleConfigChange: () => {}, handleSaveSettings: () => {}, fetchAudioCaps: () => {},
    toggleSRT: () => {}, toggleIcecast: () => {}, toggleRecording,
    updateRecording: () => {}, toggleProtectRecording: () => {}, deleteRecording: () => {},
  };
};
