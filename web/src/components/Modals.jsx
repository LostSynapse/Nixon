// web/src/components/Modals.jsx
import React, { useState, useEffect, useCallback } from 'react';
import { Settings, X, Wifi, Mic, Disc, Rss, Network, Check, Copy, ArrowRight } from 'lucide-react';
import { toast } from 'sonner';

// ... (Toast and other components remain the same) ...

// Main Settings Modal
export function SettingsModal({ show, onClose, config, onConfigChange, onSave }) {
  const [activeTab, setActiveTab] = useState('system');
  const [fullConfig, setFullConfig] = useState(config);
  const [deviceCaps, setDeviceCaps] = useState({ rates: [], depths: [] });
  const [capsLoading, setCapsLoading] = useState(false);
  const [capsError, setCapsError] = useState('');

  // Sync internal state with prop
  useEffect(() => {
    setFullConfig(config);
  }, [config, show]);

  // Debounce API call for audio caps
  const fetchAudioCaps = useCallback((device) => {
    setCapsLoading(true);
    setCapsError('');
    fetch(`/api/system/audiocaps?device=${encodeURIComponent(device)}`)
      .then(res => {
        if (!res.ok) {
          throw new Error('Failed to fetch capabilities from server.');
        }
        return res.json();
      })
      .then(data => {
        setDeviceCaps(data);
        setCapsLoading(false);
      })
      .catch(err => {
        setCapsError(err.message);
        setDeviceCaps({ rates: [], depths: [] });
        setCapsLoading(false);
      });
  }, []);

  // Fetch caps when modal opens or device changes
  useEffect(() => {
    if (show && fullConfig.audio_settings?.device) {
      fetchAudioCaps(fullConfig.audio_settings.device);
    }
  }, [show, fullConfig.audio_settings?.device, fetchAudioCaps]);


  const handleConfigChange = (e) => {
    const { name, value, type, checked } = e.target;
    const keys = name.split('.');
    let val = type === 'checkbox' ? checked : value;
    
    // Handle number inputs
    if (type === 'number') {
      val = parseFloat(val);
      if (isNaN(val)) val = 0;
    }

    const newConfig = { ...fullConfig };
    let current = newConfig;
    keys.forEach((key, index) => {
      if (index === keys.length - 1) {
        current[key] = val;
      } else {
        if (!current[key]) {
          current[key] = {};
        }
        current = current[key];
      }
    });
    setFullConfig(newConfig); // Update local state
    onConfigChange(newConfig); // Update parent state immediately
  };

  const handleDeviceChange = (e) => {
    handleConfigChange(e);
    // Refetch caps on device change
    fetchAudioCaps(e.target.value);
  }

  const renderTab = () => {
    switch (activeTab) {
      case 'system':
        return (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-400 mb-1">PipeWire Source Device</label>
              <input 
                type="text" 
                name="audio_settings.device"
                value={fullConfig.audio_settings?.device} 
                onChange={handleDeviceChange}
                className="w-full bg-gray-700 rounded-lg px-3 py-2" 
                placeholder="default"
              />
              {/* --- CORRECTED HELPER TEXT --- */}
              <p className="text-xs text-gray-500 mt-1">
                Run 'pw-cli ls Node' to find device <span className="font-mono text-xs">node.name</span>. Use 'default' for the system default.
              </p>
            </div>
            {capsLoading && <p className="text-sm text-blue-400">Loading capabilities...</p>}
            {capsError && <p className="text-sm text-red-400">Error: {capsError}</p>}
          </div>
        );
      case 'audio':
        return (
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-400 mb-1">Sample Rate</label>
              <select
                name="audio_settings.sample_rate"
                value={fullConfig.audio_settings?.sample_rate}
                onChange={handleConfigChange}
                className="w-full bg-gray-700 rounded-lg px-3 py-2"
                disabled={capsLoading || !!capsError}
              >
                {[...new Set([44100, 48000, 96000, ...deviceCaps.rates.map(r => parseInt(r, 10))])].filter(r => r > 0).sort((a,b) => a-b).map(rate => (
                  <option key={rate} value={rate} disabled={deviceCaps.rates.length > 0 && !deviceCaps.rates.includes(String(rate))}>
                    {rate / 1000} kHz
                  </option>
                 ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-400 mb-1">Bit Depth</label>
              <select
                name="audio_settings.bit_depth"
                value={fullConfig.audio_settings?.bit_depth}
                onChange={handleConfigChange}
                className="w-full bg-gray-700 rounded-lg px-3 py-2"
                disabled={capsLoading || !!capsError}
              >
                 {[...new Set([16, 24, 32, ...deviceCaps.depths.map(d => parseInt(d, 10))])].filter(d => d > 0).sort((a,b) => a-b).map(depth => (
                  <option key={depth} value={depth} disabled={deviceCaps.depths.length > 0 && !deviceCaps.depths.includes(String(depth))}>
                    {depth}-bit
                  </option>
                 ))}
              </select>
            </div>
             <div>
              <label className="block text-sm font-medium text-gray-400 mb-1">Master Channels (e.g., 1,2)</label>
              <input 
                type="text" 
                name="audio_settings.master_channels_str" 
                value={fullConfig.audio_settings?.master_channels?.join(',') || '1,2'}
                onChange={(e) => {
                  const val = e.target.value.split(',').map(s => parseInt(s.trim(), 10)).filter(n => !isNaN(n) && n > 0);
                  const newConfig = { ...fullConfig, audio_settings: { ...fullConfig.audio_settings, master_channels: val }};
                  setFullConfig(newConfig);
                  onConfigChange(newConfig);
                }}
                className="w-full bg-gray-700 rounded-lg px-3 py-2"
              />
            </div>
          </div>
        );
      case 'auto_record':
        return (
          <div className="space-y-4">
             <label className="flex items-center space-x-3">
              <input
                type="checkbox"
                name="auto_record.enabled"
                checked={fullConfig.auto_record?.enabled}
                onChange={handleConfigChange}
                className="rounded bg-gray-700 border-gray-600 text-blue-500 focus:ring-blue-600"
              />
              <span className="text-sm font-medium text-gray-300">Enable Auto-Recording (VAD)</span>
            </label>
             <label className="flex items-center space-x-3">
              <input
                type="checkbox"
                name="auto_record.smart_split_enabled"
                checked={fullConfig.auto_record?.smart_split_enabled}
                onChange={handleConfigChange}
                className="rounded bg-gray-700 border-gray-600 text-blue-500 focus:ring-blue-600"
              />
              <span className="text-sm font-medium text-gray-300">Enable Smart Split</span>
            </label>
            <div>
              <label className="block text-sm font-medium text-gray-400 mb-1">Recording Directory</label>
              <input 
                type="text" 
                name="auto_record.directory"
                value={fullConfig.auto_record?.directory} 
                onChange={handleConfigChange}
                className="w-full bg-gray-700 rounded-lg px-3 py-2"
              />
            </div>
            <div className="grid grid-cols-2 gap-4">
               <div>
                <label className="block text-sm font-medium text-gray-400 mb-1">Pre-Roll (seconds)</label>
                <input 
                  type="number" 
                  name="auto_record.preroll_duration"
                  value={fullConfig.auto_record?.preroll_duration} 
                  onChange={handleConfigChange}
                  className="w-full bg-gray-700 rounded-lg px-3 py-2"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-400 mb-1">Silence Timeout (seconds)</label>
                <input 
                  type="number" 
                  name="auto_record.smart_split_timeout"
                  value={fullConfig.auto_record?.smart_split_timeout} 
                  onChange={handleConfigChange}
                  className="w-full bg-gray-700 rounded-lg px-3 py-2"
                />
              </div>
            </div>
             <div>
              <label className="block text-sm font-medium text-gray-400 mb-1">VAD Trigger Level (dB)</label>
              <input 
                  type="number" 
                  name="auto_record.vad_db_threshold"
                  value={fullConfig.auto_record?.vad_db_threshold} 
                  onChange={handleConfigChange}
                  className="w-full bg-gray-700 rounded-lg px-3 py-2" 
                  step="1.0"
              />
              <p className="text-xs text-gray-500 mt-1">Audio below this level (e.g., -50.0) is considered silence. A higher value (e.g., -30.0) requires a louder signal.</p>
            </div>
          </div>
        );
      case 'icecast':
        return (
           <div className="space-y-4">
             <label className="flex items-center space-x-3">
              <input
                type="checkbox"
                name="icecast_settings.icecast_enabled"
                checked={fullConfig.icecast_settings?.icecast_enabled}
                onChange={handleConfigChange}
                className="rounded bg-gray-700 border-gray-600 text-blue-500 focus:ring-blue-600"
              />
              <span className="text-sm font-medium text-gray-300">Enable Icecast Stream</span>
            </label>
             <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-400 mb-1">Host</label>
                <input 
                  type="text" 
                  name="icecast_settings.icecast_host"
                  value={fullConfig.icecast_settings?.icecast_host} 
                  onChange={handleConfigChange}
                  className="w-full bg-gray-700 rounded-lg px-3 py-2"
                />
              </div>
               <div>
                <label className="block text-sm font-medium text-gray-400 mb-1">Port</label>
                <input 
                  type="number" 
                  name="icecast_settings.icecast_port"
                  value={fullConfig.icecast_settings?.icecast_port} 
                  onChange={handleConfigChange}
                  className="w-full bg-gray-700 rounded-lg px-3 py-2"
                />
              </div>
            </div>
             <div>
              <label className="block text-sm font-medium text-gray-400 mb-1">Mount Point</label>
              <input 
                type="text" 
                name="icecast_settings.icecast_mount"
                value={fullConfig.icecast_settings?.icecast_mount} 
                onChange={handleConfigChange}
                className="w-full bg-gray-700 rounded-lg px-3 py-2"
                placeholder="/stream"
              />
            </div>
             <div>
              <label className="block text-sm font-medium text-gray-400 mb-1">Password</label>
              <input 
                type="password" 
                name="icecast_settings.icecast_password"
                value={fullConfig.icecast_settings?.icecast_password} 
                onChange={handleConfigChange}
                className="w-full bg-gray-700 rounded-lg px-3 py-2"
              />
            </div>
             <div>
              <label className="block text-sm font-medium text-gray-400 mb-1">Encoding Bitrate (kbps)</label>
              <select
                name="icecast_settings.icecast_bitrate"
                value={fullConfig.icecast_settings?.icecast_bitrate}
                onChange={handleConfigChange}
                className="w-full bg-gray-700 rounded-lg px-3 py-2"
              >
                {[128000, 192000, 256000, 320000].map(rate => (
                  <option key={rate} value={rate}>{rate / 1000} kbps</option>
                 ))}
              </select>
            </div>
          </div>
        );
      case 'srt':
         return (
           <div className="space-y-4">
             <label className="flex items-center space-x-3">
              <input
                type="checkbox"
                name="srt_settings.srt_enabled"
                checked={fullConfig.srt_settings?.srt_enabled}
                onChange={handleConfigChange}
                className="rounded bg-gray-700 border-gray-600 text-blue-500 focus:ring-blue-600"
              />
              <span className="text-sm font-medium text-gray-300">Enable SRT Stream</span>
            </label>
             <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-400 mb-1">Host</label>
                <input 
                  type="text" 
                  name="srt_settings.srt_host"
                  value={fullConfig.srt_settings?.srt_host} 
                  onChange={handleConfigChange}
                  className="w-full bg-gray-700 rounded-lg px-3 py-2"
                />
              </div>
               <div>
                <label className="block text-sm font-medium text-gray-400 mb-1">Port</Mabel>
                <input 
                  type="number" 
                  name="srt_settings.srt_port"
                  value={fullConfig.srt_settings?.srt_port} 
                  onChange={handleConfigChange}
                  className="w-full bg-gray-700 rounded-lg px-3 py-2"
                />
              </div>
            </div>
             <div>
              <label className="block text-sm font-medium text-gray-400 mb-1">Encoding Bitrate (bps)</label>
              <select
                name="srt_settings.srt_bitrate"
                value={fullConfig.srt_settings?.srt_bitrate}
                onChange={handleConfigChange}
                className="w-full bg-gray-700 rounded-lg px-3 py-2"
              >
                {[64000, 96000, 128000, 256000].map(rate => (
                  <option key={rate} value={rate}>{rate / 1000} kbps</option>
                 ))}
              </select>
            </div>
          </div>
        );
      case 'network':
         return (
           <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-400 mb-1">Signaling Server URL</label>
                <input 
                  type="text" 
                  name="network_settings.signaling_url"
                  value={fullConfig.network_settings?.signaling_url} 
                  onChange={handleConfigChange}
                  className="w-full bg-gray-700 rounded-lg px-3 py-2"
                  placeholder="wss://your-signaling-server.com"
                />
              </div>
               <div>
                <label className="block text-sm font-medium text-gray-400 mb-1">STUN Server URL</label>
                <input 
                  type="text" 
                  name="network_settings.stun_url"
                  value={fullConfig.network_settings?.stun_url} 
                  onChange={handleConfigChange}
                  className="w-full bg-gray-700 rounded-lg px-3 py-2"
                />
              </div>
          </div>
        );
      default:
        return null;
    }
  };

  if (!show) {
    return null;
  }

  const tabClass = (tabName) => 
    `flex-1 px-3 py-2 text-sm font-medium rounded-lg text-center cursor-pointer ${
      activeTab === tabName 
        ? 'bg-blue-600 text-white' 
        : 'text-gray-400 hover:bg-gray-700'
    }`;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-75 flex items-center justify-center z-50 p-4">
      <div className="bg-gray-800 rounded-lg shadow-2xl w-full max-w-2xl border border-gray-700">
        <div className="flex justify-between items-center p-4 border-b border-gray-700">
          <h2 className="text-lg font-semibold text-white">Settings</h2>
          <button onClick={onClose} className="text-gray-400 hover:text-white">
            <X size={20} />
          </button>
        </div>
        
        <div className="flex p-4 space-x-2">
          {/* Sidebar */}
          <nav className="w-1/4 space-y-1">
             <button onClick={() => setActiveTab('system')} className={tabClass('system')}>
               <div className="flex items-center"><Mic size={16} className="mr-2" />System</div>
            </button>
             <button onClick={() => setActiveTab('audio')} className={tabClass('audio')}>
               <div className="flex items-center"><Disc size={16} className="mr-2" />Audio</div>
            </button>
             <button onClick={() => setActiveTab('auto_record')} className={tabClass('auto_record')}>
               <div className="flex items-center"><Rss size={16} className="mr-2" />Auto-Record</div>
            </button>
             <button onClick={() => setActiveTab('icecast')} className={tabClass('icecast')}>
               <div className="flex items-center"><Wifi size={16} className="mr-2" />Icecast</div>
            </button>
             <button onClick={() => setActiveTab('srt')} className={tabClass('srt')}>
                <div className="flex items-center"><ArrowRight size={16} className="mr-2" />SRT</div>
            </button>
             <button onClick={() => setActiveTab('network')} className={tabClass('network')}>
               <div className="flex items-center"><Network size={16} className="mr-2" />Network</div>
            </button>
          </nav>
          
          {/* Content */}
          <div className="w-3/4 bg-gray-900 p-4 rounded-lg">
            {renderTab()}
          </div>
        </div>

        <div className="p-4 border-t border-gray-700 flex justify-end">
          <button 
            onClick={onSave} 
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            Save and Close
          </button>
        </div>
      </div>
    </div>
  );
}

