// web/src/App.jsx
import React, { useState, useCallback, useRef, useEffect } from 'react';
import { Settings, Wifi, WifiOff } from 'lucide-react';
import { SettingsModal, ConfirmationModal, EditModal } from './components/Modals';
import StreamControl from './components/StreamControl';
import RecordingControl from './components/RecordingControl';
import Logo from './components/Logo';
import DiskUsage from './components/DiskUsage';
import RecordingsList from './components/RecordingsList';
import { useNixonApi } from './hooks/useNixonApi';

export default function App() {
  const [modal, setModal] = useState(null);
  const { appStatus, config, recordings, isConnected, toggleSRT, toggleIcecast, toggleRecording, handleConfigChange, handleSaveSettings, fetchAudioCaps, audioCaps, updateRecording, toggleProtectRecording, deleteRecording } = useNixonApi();

  const openSettingsModal = useCallback(() => { setModal({ type: 'settings' }); if (config?.audio_settings?.device) { fetchAudioCaps(config.audio_settings.device); } }, [config, fetchAudioCaps]);
  const openEditModal = useCallback((recording) => setModal({ type: 'edit', data: recording }), []);
  const openDeleteModal = useCallback((recording) => setModal({ type: 'delete', data: recording }), []);
  const openUnlockModal = useCallback((recording) => setModal({ type: 'unlock', data: recording }), []);
  const closeModal = useCallback(() => setModal(null), []);

  const onSaveSettings = useCallback(async () => { const success = await handleSaveSettings(); if (success) { closeModal(); } else { alert("Failed to save settings."); } }, [handleSaveSettings, closeModal]);
  const onSaveEdit = useCallback(async (editedRecording) => { const success = await updateRecording(editedRecording.ID, editedRecording.Name, editedRecording.Notes, editedRecording.Genre); if (success) closeModal(); else alert("Failed to save changes."); }, [updateRecording, closeModal]); // Use GORM fields
  const onConfirmDelete = useCallback(async () => { if (modal?.type === 'delete' && modal.data?.ID) { const success = await deleteRecording(modal.data.ID); if (success) closeModal(); else alert("Failed to delete."); } }, [modal, deleteRecording, closeModal]); // Use GORM ID
  const onConfirmUnlock = useCallback(async () => { if (modal?.type === 'unlock' && modal.data?.ID) { const success = await toggleProtectRecording(modal.data.ID); if (success) closeModal(); else alert("Failed to unlock."); } }, [modal, toggleProtectRecording, closeModal]); // Use GORM ID

  const [playingFile, setPlayingFile] = useState(null);
  const [isAudioPlaying, setIsAudioPlaying] = useState(false);
  const audioRef = useRef(null);
  const handlePlayPause = useCallback((filename) => { const newSrc = "/recordings/" + filename; if (playingFile === newSrc && isAudioPlaying) { audioRef.current?.pause(); } else { setPlayingFile(newSrc); } }, [playingFile, isAudioPlaying]);

  useEffect(() => {
    const audioEl = audioRef.current; if (!audioEl) return;
    const handlePlay = () => setIsAudioPlaying(true); const handlePause = () => setIsAudioPlaying(false); const handleEnded = () => { setIsAudioPlaying(false); setPlayingFile(null); };
    audioEl.addEventListener('play', handlePlay); audioEl.addEventListener('pause', handlePause); audioEl.addEventListener('ended', handleEnded);
    if (playingFile) { audioEl.play().catch(e => { console.error("Audio play failed:", e); setIsAudioPlaying(false); setPlayingFile(null); }); } else { audioEl.pause(); }
    return () => { audioEl.removeEventListener('play', handlePlay); audioEl.removeEventListener('pause', handlePause); audioEl.removeEventListener('ended', handleEnded); };
  }, [playingFile]);

  if (!config) { return (<div className="bg-gray-900 text-white min-h-screen flex items-center justify-center p-4"><div className="flex flex-col items-center text-center"><Logo className="w-16 h-16 text-gray-600 animate-pulse" /><div className="text-lg text-gray-500 mt-4 mb-2">Loading...</div>{!isConnected && <div className="text-sm text-red-500">(Connecting...)</div>}</div></div>); }

  return (
    <div className="bg-gray-900 text-white min-h-screen font-sans p-4 md:p-6 pb-24">
      {modal?.type === 'settings' && <SettingsModal fullConfig={config} audioCaps={audioCaps} onConfigChange={handleConfigChange} onSave={onSaveSettings} onCancel={closeModal} onFetchCaps={fetchAudioCaps} />}
      {modal?.type === 'edit' && <EditModal recording={modal.data} onSave={onSaveEdit} onCancel={closeModal} />}
      {modal?.type === 'delete' && <ConfirmationModal title="Delete?" message={`Delete '${modal.data.Name || modal.data.Filename}'?`} onConfirm={onConfirmDelete} onCancel={closeModal} />}
      {modal?.type === 'unlock' && <ConfirmationModal title="Unlock?" message={`Unlock '${modal.data.Name || modal.data.Filename}'?`} onConfirm={onConfirmUnlock} onCancel={closeModal} />}

      <div className="max-w-7xl mx-auto">
        <header className="flex justify-between items-center mb-6">
          <div className="flex items-center"><Logo className="w-8 h-8 mr-3 text-gray-400" /><h1 className="text-2xl md:text-3xl font-bold">Nixon</h1></div>
          <div className="flex items-center space-x-4">
            <button onClick={openSettingsModal} className="text-gray-400 hover:text-white" title="Settings"><Settings size={24} /></button>
            <div className="flex items-center text-sm" title={isConnected ? 'Connected' : 'Disconnected'}>{isConnected ? <Wifi size={18} className="text-green-500" /> : <WifiOff size={18} className="text-red-500" />}</div>
          </div>
        </header>
        <main>
          <section className="mb-6"><h2 className="text-xl font-semibold border-b-2 border-gray-700 pb-2 mb-4">Live Controls</h2><div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {config.srt_settings?.srt_enabled && <StreamControl title="SRT Stream" status={appStatus.is_streaming_srt} onToggle={toggleSRT} />}
            {config.icecast_settings?.icecast_enabled && <StreamControl title="Icecast Stream" status={appStatus.is_streaming_icecast} onToggle={toggleIcecast} listeners={appStatus.listeners} peak={appStatus.listener_peak} />}
            <RecordingControl status={appStatus.is_recording} currentFile={appStatus.current_recording_file} onToggle={() => toggleRecording(!appStatus.is_recording)} onAutoToggle={() => { handleConfigChange({ target: { name: 'auto_record.enabled', type: 'checkbox', checked: !config.auto_record?.enabled }}); handleSaveSettings(); }} autoRecordEnabled={config.auto_record?.enabled} /> {/* Split button removed for simplicity, add back if needed */}
          </div></section>
          <RecordingsList recordings={recordings} onPlayPause={handlePlayPause} playingFile={playingFile} isAudioPlaying={isAudioPlaying} onEdit={openEditModal} onProtect={(rec) => rec.Protected ? openUnlockModal(rec) : toggleProtectRecording(rec.ID)} onDelete={openDeleteModal} />
        </main>
      </div>
      <DiskUsage usage={appStatus.disk_usage_percent} /> <audio ref={audioRef} src={playingFile} className="hidden" />
    </div>
  );
}

