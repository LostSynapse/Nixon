import { useState } from 'react';
import Logo from './components/Logo';
import RecordingControl from './components/RecordingControl';
import StreamControl from './components/StreamControl';
import RecordingsList from './components/RecordingsList';
import DiskUsage from './components/DiskUsage';
import { useNixonApi } from './hooks/useNixonApi';
import { SettingsModal, AudioDeviceModal } from './components/Modals';

function App() {
  const {
    appStatus, config, recordings, audioCaps, isConnected,
    handleConfigChange, handleSaveSettings, fetchAudioCaps,
    toggleSRT, toggleIcecast, toggleRecording,
    updateRecording, toggleProtectRecording, deleteRecording
  } = useNixonApi();

  const [isSettingsModalOpen, setSettingsModalOpen] = useState(false);
  const [isAudioDeviceModalOpen, setAudioDeviceModalOpen] = useState(false);

  const handleOpenAudioDeviceModal = () => {
    fetchAudioCaps(); // Fetch the latest device capabilities
    setAudioDeviceModalOpen(true);
  };

  return (
    <div className="bg-gray-900 text-white min-h-screen font-sans">
      <header className="container mx-auto px-4 py-6 flex justify-between items-center">
        <Logo isConnected={isConnected} appState={appStatus?.state} />
        <h1 className="text-4xl font-bold">Nixon</h1>
        <div className="flex items-center space-x-2">
          <button onClick={() => setSettingsModalOpen(true)} className="bg-gray-700 hover:bg-gray-600 text-white font-bold py-2 px-4 rounded">
            Settings
          </button>
          <button onClick={handleOpenAudioDeviceModal} className="bg-gray-700 hover:bg-gray-600 text-white font-bold py-2 px-4 rounded">
            Audio Devices
          </button>
        </div>
      </header>

      <main className="container mx-auto px-4">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">

          {/* Column 1: Controls */}
          <div className="space-y-8">
          <RecordingControl
            status={appStatus?.state === 'recording'}
            currentFile={appStatus?.currentRecFile}
            onToggle={toggleRecording}
            autoRecordEnabled={config?.autoRecord?.enabled}
            // onSplit and onAutoToggle are not implemented in the hook yet, so we pass placeholders
            onSplit={() => console.log('Split function not implemented')}
            onAutoToggle={() => console.log('Auto-toggle function not implemented')}
          />


          <StreamControl
            title="Icecast Stream"
            status={config?.icecast?.enabled}
            onToggle={toggleIcecast}
            // Listeners and peak are not yet available in the data model, so pass placeholders.
            listeners={0} 
            peak={0}
          />
          <StreamControl
           title="SRT Stream"
            status={config?.srt?.enabled}
            onToggle={toggleSRT}
            />
          </div>

          {/* Column 2: Recordings */}
          <div className="lg:col-span-2 space-y-8">
            <RecordingsList
              recordings={recordings}
              onUpdate={updateRecording}
              onDelete={deleteRecording}
              onToggleProtect={toggleProtectRecording}
            />
          </div>

          {/* Column 3 (becomes footer on small screens): System Stats */}
          <div className="md:col-span-2 lg:col-span-3 space-y-4">
            <h2 className="text-2xl font-semibold border-b-2 border-gray-700 pb-2">System Status</h2>
            {!isConnected && <p className="text-red-500">Connecting to server...</p>}
            {isConnected && appStatus && (
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                <DiskUsage used={appStatus?.diskUsage} />
                <div className="bg-gray-800 p-4 rounded-lg text-center">
                  <h3 className="text-lg font-medium">CPU Usage</h3>
                  <p className="text-3xl font-bold">{appStatus?.cpuUsage?.toFixed(2) ?? 'N/A'}%</p>
                </div>
                <div className="bg-gray-800 p-4 rounded-lg text-center">
                  <h3 className="text-lg font-medium">Memory Usage</h3>
                  <p className="text-3xl font-bold">{appStatus?.memoryUsage?.toFixed(2) ?? 'N/A'}%</p>
                </div>
              </div>
            )}
          </div>
        </div>
      </main>

      {isSettingsModalOpen && (
        <SettingsModal
          config={config}
          onConfigChange={handleConfigChange}
          onSave={handleSaveSettings}
          onClose={() => setSettingsModalOpen(false)}
        />
      )}

      {isAudioDeviceModalOpen && (
        <AudioDeviceModal
          audioCaps={audioCaps}
          onClose={() => setAudioDeviceModalOpen(false)}
        />
      )}


      <footer className="container mx-auto px-4 py-6 text-center text-gray-500">
        <p>&copy; 2024 Nixon. All Rights Reserved.</p>
      </footer>
    </div>
  );
}

export default App;
