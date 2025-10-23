// web/src/hooks/useNixonApi.js
import { useState, useEffect, useRef, useCallback } from 'react';

export function useNixonApi() {
  const [appStatus, setAppStatus] = useState({ disk_usage_percent: 0 });
  const [config, setConfig] = useState(null);
  const [recordings, setRecordings] = useState(null);
  const [audioCaps, setAudioCaps] = useState({ rates: [], depths: [] });
  const [isConnected, setIsConnected] = useState(false);
  const socket = useRef(null);

  const handleApiCall = useCallback(async (endpoint, method = 'POST', body = null) => {
    try {
      const options = { method, headers: body ? { 'Content-Type': 'application/json' } : {} };
      if (body) options.body = JSON.stringify(body);
      const response = await fetch(`/api/${endpoint}`, options);
      if (!response.ok) { console.error("API Error:", response.status, await response.text()); return false; }
      return true;
    } catch (error) { console.error('API call fetch error:', error); return false; }
  }, []);

  const fetchRecordings = useCallback(async () => {
    try {
      const response = await fetch('/api/recordings');
      if (response.ok) { setRecordings(await response.json()); }
      else { console.error("Failed fetch recordings:", response.status); setRecordings([]); }
    } catch (error) { console.error("Failed fetch recordings:", error); setRecordings([]); }
  }, []);

  const fetchConfig = useCallback(async () => {
    try {
      const response = await fetch('/api/config');
      if (response.ok) {
        const newConfig = await response.json();
        if (!newConfig.audio_settings) newConfig.audio_settings = {};
        if (!Array.isArray(newConfig.audio_settings.master_channels)) newConfig.audio_settings.master_channels = [1, 2];
        if (!newConfig.auto_record) newConfig.auto_record = {};
        if (!newConfig.srt_settings) newConfig.srt_settings = {};
        if (!newConfig.icecast_settings) newConfig.icecast_settings = {};
        if (!newConfig.network_settings) newConfig.network_settings = {};
        setConfig(newConfig);
      } else { console.error("Failed fetch config:", response.status); }
    } catch (error) { console.error("Failed fetch config:", error); }
  }, []);

  const fetchAudioCaps = useCallback(async (device) => {
    if (!device || device === "default" || device === "") { setAudioCaps({ rates: [], depths: [] }); return; }
    try {
      const response = await fetch(`/api/capabilities?device=${encodeURIComponent(device)}`);
      if (response.ok) { setAudioCaps(await response.json()); }
      else { console.error("Failed fetch audio caps:", response.status); setAudioCaps({ rates: [], depths: [] }); }
    } catch (error) { console.error("Failed fetch audio caps:", error); setAudioCaps({ rates: [], depths: [] }); }
  }, []);

  useEffect(() => {
    function connect() {
      const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const wsUrl = `${wsProtocol}//${window.location.host}/ws`;
      socket.current = new WebSocket(wsUrl);
      socket.current.onopen = () => { console.log("WS Connected"); setIsConnected(true); fetchConfig(); fetchRecordings(); }
      socket.current.onclose = () => { console.log("WS Disconnected"); setIsConnected(false); setTimeout(connect, 5000); };
      socket.current.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          // console.log("WS Message:", data); // Reduce console noise
          if (data.audio_settings) { // Config update heuristic
            setConfig(prev => ({ ...prev, ...data }));
          } else { // Status update
            setAppStatus(prev => { if (prev.is_recording && data.is_recording === false) { fetchRecordings(); } return data; });
          }
        } catch (e) { console.error("Failed parse WS message:", e, event.data); }
      };
      socket.current.onerror = (err) => { console.error("WS Error:", err); socket.current.close(); };
    }
    connect();
    return () => { console.log("Closing WS connection."); socket.current?.close(); }
  }, [fetchRecordings, fetchConfig]);

  useEffect(() => { fetchConfig(); fetchRecordings(); }, [fetchConfig, fetchRecordings]);

  const handleConfigChange = useCallback((e) => {
    const { name, value, type, checked } = e.target;
    const keys = name.split('.');
    setConfig(prev => {
      if (!prev) return null;
      const newConfig = JSON.parse(JSON.stringify(prev));
      let current = newConfig;
      for (let i = 0; i < keys.length - 1; i++) {
        if (current[keys[i]] === undefined || current[keys[i]] === null) { current[keys[i]] = {}; }
        current = current[keys[i]];
      }
      let finalValue; const dataType = e.target.dataset.type;
      if (type === 'checkbox') { finalValue = checked; }
      else if (type === 'number' || dataType === 'int' || dataType === 'float') {
        finalValue = parseFloat(value); if (isNaN(finalValue)) { finalValue = dataType === 'float' ? 0.0 : 0; }
        if (dataType === 'int' && !isNaN(finalValue)) { finalValue = Math.round(finalValue); }
      } else if (dataType === 'int_array') {
        finalValue = value.split(',').map(s => parseInt(s.trim(), 10)).filter(n => !isNaN(n) && n > 0);
        if (finalValue.length === 0) finalValue = [1, 2];
      } else { finalValue = value; }
      current[keys[keys.length - 1]] = finalValue; return newConfig;
    });
    if (name === 'audio_settings.device') { fetchAudioCaps(value); }
  }, [fetchAudioCaps]);

  const handleSaveSettings = useCallback(async () => {
    if (!config) return false;
    const configToSave = JSON.parse(JSON.stringify(config));
    if (configToSave.audio_settings && typeof configToSave.audio_settings.master_channels === 'string') {
      configToSave.audio_settings.master_channels = configToSave.audio_settings.master_channels.split(',').map(s => parseInt(s.trim(), 10)).filter(n => !isNaN(n) && n > 0);
      if (configToSave.audio_settings.master_channels.length === 0) configToSave.audio_settings.master_channels = [1, 2];
    }
    const success = await handleApiCall('config', 'POST', configToSave);
    if (success) { fetchConfig(); } return success;
  }, [config, handleApiCall, fetchConfig]);

  const toggleSRT = useCallback(() => { if (!config || !config.srt_settings) return; handleApiCall('stream', 'POST', { stream: 'srt', enabled: !appStatus.is_streaming_srt }) }, [handleApiCall, appStatus.is_streaming_srt, config]);
  const toggleIcecast = useCallback(() => { if (!config || !config.icecast_settings) return; handleApiCall('stream', 'POST', { stream: 'icecast', enabled: !appStatus.is_streaming_icecast }) }, [handleApiCall, appStatus.is_streaming_icecast, config]);
  const toggleRecording = useCallback((enable) => { handleApiCall(enable ? 'record/start' : 'record/stop', 'POST'); }, [handleApiCall]);
  const updateRecording = useCallback(async (id, name, notes, genre) => { const success = await handleApiCall(`recordings/${String(id)}`, 'PUT', { name, notes, genre }); if (success) fetchRecordings(); return success; }, [handleApiCall, fetchRecordings]);
  const toggleProtectRecording = useCallback(async (id) => { const success = await handleApiCall(`recordings/${String(id)}/protect`, 'POST'); if (success) fetchRecordings(); return success; }, [handleApiCall, fetchRecordings]);
  const deleteRecording = useCallback(async (id) => { const success = await handleApiCall(`recordings/${String(id)}`, 'DELETE'); if (success) fetchRecordings(); return success; }, [handleApiCall, fetchRecordings]);

  return { appStatus, config, recordings, audioCaps, isConnected, handleConfigChange, handleSaveSettings, fetchAudioCaps, toggleSRT, toggleIcecast, toggleRecording, updateRecording, toggleProtectRecording, deleteRecording };
}

