import React, { useState, useEffect, useRef } from 'react';
import { Power, Mic, Waves, AlertTriangle, Play, Download, Trash2, Lock, Unlock, Edit, Settings, X, Pause, Users, ArrowUp } from 'lucide-react';
import { Modal, ConfirmationModal, EditModal, SettingsModal } from './components/Modals.jsx';
import StreamControl from './components/StreamControl.jsx';
import RecordingControl from './components/RecordingControl.jsx';
import DiskUsage from './components/DiskUsage.jsx';
import RecordingsList from './components/RecordingsList.jsx';

export default function App() {
    const [appStatus, setAppStatus] = useState({ DiskUsagePercent: 0, Listeners: 0, ListenerPeak: 0 });
    const [recordings, setRecordings] = useState(null);
    const [isConnected, setIsConnected] = useState(false);
    const [playingFile, setPlayingFile] = useState(null);
    const [isAudioPlaying, setIsAudioPlaying] = useState(false);
    const [modal, setModal] = useState(null);
    const [systemConfig, setSystemConfig] = useState({ srt_enabled: true, icecast_enabled: true, auto_record: { enabled: false } });
    const audioRef = useRef(null);
    const socket = useRef(null);

    const fetchSystemConfig = async () => { try { const response = await fetch('/api/settings/all'); if (response.ok) setSystemConfig(await response.json()); } catch (error) { console.error("Failed to fetch system config:", error); } };
    const fetchRecordings = async () => { try { const response = await fetch('/api/recordings'); if (response.ok) setRecordings(await response.json()); } catch (error) { console.error("Failed to fetch recordings:", error); } };

    useEffect(() => {
        fetchRecordings();
        fetchSystemConfig();
        function connect() {
            const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = `${wsProtocol}//${window.location.host}/ws`;
            socket.current = new WebSocket(wsUrl);
            socket.current.onopen = () => setIsConnected(true);
            socket.current.onclose = () => { setIsConnected(false); setTimeout(connect, 3000); };
            socket.current.onmessage = (event) => {
                const data = JSON.parse(event.data);
                setAppStatus(prev => {
                    if (prev.RecordingActive && !data.RecordingActive) { setTimeout(fetchRecordings, 200); }
                    return {...prev, ...data};
                });
            };
            socket.current.onerror = (err) => { console.error("WS Error:", err); socket.current.close(); };
        }
        connect();
        return () => socket.current?.close();
    }, []);

    const handleAPICall = async (endpoint, action, method = 'POST', body = null) => {
        try {
            const options = { method, headers: body ? { 'Content-Type': 'application/json' } : {} };
            if (body) options.body = JSON.stringify(body);
            const response = await fetch(`/api/${endpoint}${action ? `/${action}`: ''}`, options);
            if (!response.ok) { console.error("API Error:", response.status, await response.text()); return false; }
            return true;
        } catch (error) { console.error('API call fetch error:', error); }
        return false;
    };

    const toggleSRT = () => handleAPICall('stream/srt', appStatus.SRTStreamActive ? 'stop' : 'start');
    const toggleIcecast = () => handleAPICall('stream/icecast', appStatus.IcecastStreamActive ? 'stop' : 'start');
    const startAllStreams = () => handleAPICall('stream/all', 'start');
    const stopAllStreams = () => handleAPICall('stream/all', 'stop');

    const toggleRecording = async (action) => { 
        const success = await handleAPICall('recording', action); 
        if (success && action === 'stop') {
             setTimeout(fetchRecordings, 500); 
        }
    };
    const splitRecording = async () => { const success = await handleAPICall('recording', 'split'); if (success) setTimeout(fetchRecordings, 500); };
    const handlePlayPause = (filename) => { const newSrc = `/recordings/${filename}`; if (playingFile === newSrc && isAudioPlaying) { audioRef.current?.pause(); } else { setPlayingFile(newSrc); } };
    const handleDelete = async (rec) => { const success = await handleAPICall('recordings', rec.id, 'DELETE'); if (success) { fetchRecordings(); setModal(null); } };
    const handleProtect = async (rec) => { const success = await handleAPICall('recordings', `${rec.id}/protect`); if (success) { fetchRecordings(); setModal(null); } };
    const handleSaveEdit = async (rec) => { const success = await handleAPICall('recordings', rec.id, 'PUT', rec); if (success) { fetchRecordings(); setModal(null); } };
    const handleSaveSettings = async (type, settings) => { const success = await handleAPICall(`settings/${type}`, '', 'POST', settings); if (success) { setModal(null); fetchSystemConfig(); alert("Settings saved. A backend restart may be required for some changes to take effect."); } };
    const handleAutoRecordToggle = () => {
        const newSettings = { ...systemConfig.auto_record, enabled: !systemConfig.auto_record.enabled };
        handleSaveSettings('system', { srt_enabled: systemConfig.srt_enabled, icecast_enabled: systemConfig.icecast_enabled, auto_record: newSettings });
    };

    useEffect(() => { if (playingFile && audioRef.current) { audioRef.current.play().catch(e => console.error("Audio play failed:", e)); } }, [playingFile]);
    useEffect(() => {
        const audioEl = audioRef.current;
        if(audioEl) {
            const onPlay = () => setIsAudioPlaying(true);
            const onPause = () => setIsAudioPlaying(false);
            const onEnded = () => { setIsAudioPlaying(false); setPlayingFile(null); };
            audioEl.addEventListener('play', onPlay);
            audioEl.addEventListener('pause', onPause);
            audioEl.addEventListener('ended', onEnded);
            return () => { audioEl.removeEventListener('play', onPlay); audioEl.removeEventListener('pause', onPause); audioEl.removeEventListener('ended', onEnded); };
        }
    }, [playingFile]);

    const numEnabledStreams = [systemConfig.srt_enabled, systemConfig.icecast_enabled].filter(Boolean).length;

    return (
        <div className="bg-gray-900 text-white min-h-screen font-sans p-4 md:p-6 pb-24">
            {modal?.type === 'edit' && <EditModal recording={modal.data} onSave={handleSaveEdit} onCancel={() => setModal(null)} />}
            {modal?.type === 'delete' && <ConfirmationModal title="Delete Recording?" message={`Are you sure you want to delete '${modal.data.name || modal.data.filename}'? This cannot be undone.`} onConfirm={() => handleDelete(modal.data)} onCancel={() => setModal(null)} />}
            {modal?.type === 'unlock' && <ConfirmationModal title="Unlock Recording?" message={`Unlock '${modal.data.name || modal.data.filename}'? It can then be deleted.`} onConfirm={() => handleProtect(modal.data)} onCancel={() => setModal(null)} />}
            {modal?.type === 'settings' && <SettingsModal onSave={{ icecast: (s) => handleSaveSettings('icecast', s), system: (s) => handleSaveSettings('system', s), audio: (s) => handleSaveSettings('audio', s) }} onCancel={() => setModal(null)} />}
            
            <div className="max-w-7xl mx-auto">
                 <header className="flex justify-between items-center mb-6"><div className="flex items-center"><img src="/nixon_logo.svg" alt="Nixon Logo" className="w-8 h-8 mr-3 text-gray-400" /><h1 className="text-2xl md:text-3xl font-bold">Nixon</h1></div><div className="flex items-center space-x-4"><button onClick={() => setModal({type: 'settings'})} className="text-gray-400 hover:text-white"><Settings size={24}/></button><div className="flex items-center"><div className={`w-3 h-3 rounded-full mr-2 ${isConnected ? 'bg-green-500' : 'bg-red-500'}`}></div><span>{isConnected ? 'Connected' : 'Disconnected'}</span></div></div></header>
                <main>
                    <section className="mb-6">
                        <div className="flex justify-between items-center border-b-2 border-gray-700 pb-2 mb-4"><h2 className="text-xl font-semibold">Live Controls</h2>{numEnabledStreams > 1 && (<div className="flex space-x-2"><button onClick={startAllStreams} className="px-3 py-1 text-sm rounded-md bg-green-600 hover:bg-green-700">Start All</button><button onClick={stopAllStreams} className="px-3 py-1 text-sm rounded-md bg-red-600 hover:bg-red-700">Stop All</button></div>)}</div>
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                            {systemConfig.srt_enabled && <StreamControl title="SRT Stream" status={appStatus.SRTStreamActive} onToggle={toggleSRT} />}
                            {systemConfig.icecast_enabled && <StreamControl title="Icecast Stream" status={appStatus.IcecastStreamActive} onToggle={toggleIcecast} listeners={appStatus.Listeners} peak={appStatus.ListenerPeak} />}
                            <RecordingControl status={appStatus.RecordingActive} currentFile={appStatus.CurrentRecordingFile} onToggle={toggleRecording} onSplit={splitRecording} onAutoToggle={handleAutoRecordToggle} autoRecordEnabled={systemConfig.auto_record.enabled} />
                        </div>
                    </section>
                    <RecordingsList recordings={recordings} onPlayPause={handlePlayPause} playingFile={playingFile} isAudioPlaying={isAudioPlaying} onEdit={(rec) => setModal({ type: 'edit', data: rec })} onProtect={(rec) => rec.protected ? setModal({ type: 'unlock', data: rec }) : handleProtect(rec)} onDelete={(rec) => setModal({ type: 'delete', data: rec })} />
                </main>
            </div>
            <DiskUsage usage={appStatus.DiskUsagePercent} />
            <audio ref={audioRef} src={playingFile} className="hidden" />
        </div>
    );
}

