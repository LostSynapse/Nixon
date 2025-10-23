// web/src/App.jsx
import React, { useState, useEffect } from 'react';
import { SettingsModal } from './components/Modals';
import StreamControl from './components/StreamControl';
import Logo from './components/Logo';

function App() {
  const [showSettings, setShowSettings] = useState(false);
  const [config, setConfig] = useState(null);

  useEffect(() => {
    // Fetch initial config from the server
    const fetchConfig = async () => {
      try {
        const response = await fetch('/api/config');
        if (response.ok) {
          const data = await response.json();
          setConfig(data);
        } else {
          console.error("Failed to fetch config");
        }
      } catch (error) {
        console.error("Error fetching config:", error);
      }
    };
    fetchConfig();
  }, []);

  return (
    <div className="bg-gray-900 text-white min-h-screen font-sans">
      <div className="container mx-auto p-8">
        <header className="flex items-center justify-between mb-8">
          <div className="flex items-center space-x-4">
            <Logo className="w-12 h-12 text-gray-400" />
            <div>
              <h1 className="text-3xl font-bold">Nixon</h1>
              <p className="text-gray-400">Advanced Audio Management</p>
            </div>
          </div>
        </header>

        <main className="space-y-6">
          <StreamControl onSettingsClick={() => setShowSettings(true)} />
          {/* Other components like recording list, etc. will go here */}
        </main>
      </div>

      <SettingsModal
        show={showSettings}
        onClose={() => setShowSettings(false)}
        config={config}
        setConfig={setConfig}
      />
    </div>
  );
}

export default App;
