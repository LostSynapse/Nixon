import React, { useState, useEffect } from 'react';
import { X } from 'lucide-react';

export const Modal = ({ children, onClose, title }) => (
    <div className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-50" onClick={onClose}>
        <div className="bg-gray-800 rounded-lg shadow-2xl p-6 w-full max-w-lg m-4" onClick={e => e.stopPropagation()}>
            <div className="flex justify-between items-center mb-4">
                <h3 className="text-xl font-bold">{title}</h3>
                <button onClick={onClose} className="text-gray-500 hover:text-white"><X size={24} /></button>
            </div>
            {children}
        </div>
    </div>
);

export const ConfirmationModal = ({ title, message, onConfirm, onCancel }) => (
    <Modal onClose={onCancel} title={title}>
        <p className="text-gray-300 mb-6">{message}</p>
        <div className="flex justify-end space-x-3">
            <button onClick={onCancel} className="px-4 py-2 rounded-lg bg-gray-600 hover:bg-gray-700 font-semibold">Cancel</button>
            <button onClick={onConfirm} className="px-4 py-2 rounded-lg bg-red-600 hover:bg-red-700 font-semibold text-white">Confirm</button>
        </div>
    </Modal>
);

export const EditModal = ({ recording, onSave, onCancel }) => {
    const [name, setName] = useState(recording.name || '');
    const [notes, setNotes] = useState(recording.notes || '');
    const [genre, setGenre] = useState(recording.genre || '');

    const handleSave = () => {
        onSave({ ...recording, name, notes, genre });
    };

    return (
        <Modal onClose={onCancel} title="Edit Recording">
            <div className="space-y-4">
                <div><label className="block text-sm font-medium text-gray-400 mb-1">Name</label><input type="text" value={name} onChange={e => setName(e.target.value)} className="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500" /></div>
                <div><label className="block text-sm font-medium text-gray-400 mb-1">Notes</label><textarea value={notes} onChange={e => setNotes(e.target.value)} rows="3" className="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"></textarea></div>
                <div><label className="block text-sm font-medium text-gray-400 mb-1">Genre</label><input type="text" value={genre} onChange={e => setGenre(e.target.value)} className="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500" /></div>
            </div>
            <div className="flex justify-end space-x-3 mt-6"><button onClick={onCancel} className="px-4 py-2 rounded-lg bg-gray-600 hover:bg-gray-700 font-semibold">Cancel</button><button onClick={handleSave} className="px-4 py-2 rounded-lg bg-blue-600 hover:bg-blue-700 font-semibold text-white">Save Changes</button></div>
        </Modal>
    );
};

export const SettingsModal = ({ onSave, onCancel }) => {
    const [activeTab, setActiveTab] = useState('system');
    const [fullConfig, setFullConfig] = useState({ icecast_settings: {}, audio_settings: {}, auto_record: {} });
    const [audioDevices, setAudioDevices] = useState([]);

    useEffect(() => {
        const fetchAllSettings = async () => {
            try {
                const [configRes, devicesRes] = await Promise.all([ fetch('/api/settings/all'), fetch('/api/system/audiodevices') ]);
                if (configRes.ok) setFullConfig(await configRes.json());
                if (devicesRes.ok) setAudioDevices(await devicesRes.json());
            } catch (e) { console.error("Could not fetch settings", e) }
        };
        fetchAllSettings();
    }, []);

    const handleConfigChange = (e) => {
        const { name, value, type, checked } = e.target;
        const keys = name.split('.');
        setFullConfig(prev => {
            let current = { ...prev };
            let final = current;
            for (let i = 0; i < keys.length - 1; i++) {
                final = final[keys[i]];
            }
            final[keys[keys.length - 1]] = type === 'checkbox' ? checked : (e.target.tagName === 'SELECT' && !isNaN(parseInt(value))) ? parseInt(value) : value;
            return current;
        });
    };

    const TabButton = ({ tabName, label }) => (<button onClick={() => setActiveTab(tabName)} className={`px-4 py-2 font-semibold transition-colors ${activeTab === tabName ? 'text-white border-b-2 border-blue-500' : 'text-gray-400 hover:text-white'}`}>{label}</button>);

    return (
        <Modal onClose={onCancel} title="Settings">
            <div className="border-b border-gray-700 mb-4"><TabButton tabName="system" label="System" /><TabButton tabName="audio" label="Audio" /><TabButton tabName="icecast" label="Icecast" /></div>
            {activeTab === 'system' && (
                 <div className="space-y-4">
                     <h4 className="text-lg font-semibold text-gray-200">Stream Widgets</h4>
                     <label className="flex items-center space-x-3 cursor-pointer"><input type="checkbox" name="srt_enabled" checked={fullConfig.srt_enabled} onChange={handleConfigChange} className="w-5 h-5 accent-blue-500" /><span className="text-gray-300">Enable SRT Stream Widget</span></label>
                     <label className="flex items-center space-x-3 cursor-pointer"><input type="checkbox" name="icecast_enabled" checked={fullConfig.icecast_enabled} onChange={handleConfigChange} className="w-5 h-5 accent-blue-500" /><span className="text-gray-300">Enable Icecast Stream Widget</span></label>
                     <hr className="border-gray-600"/>
                     <h4 className="text-lg font-semibold text-gray-200">Auto Recording</h4>
                     <label className="flex items-center space-x-3 cursor-pointer"><input type="checkbox" name="auto_record.enabled" checked={fullConfig.auto_record?.enabled} onChange={handleConfigChange} className="w-5 h-5 accent-blue-500" /><span className="text-gray-300">Enable Automatic Recording (VAD)</span></label>
                     <div><label className="block text-sm font-medium text-gray-400 mb-1">Silence Timeout (seconds)</label><input type="number" name="auto_record.timeout_seconds" value={fullConfig.auto_record?.timeout_seconds} onChange={handleConfigChange} className="w-full bg-gray-700 rounded-lg px-3 py-2" /></div>
                     <div className="flex justify-end pt-4"><button onClick={() => onSave.system({ srt_enabled: fullConfig.srt_enabled, icecast_enabled: fullConfig.icecast_enabled, auto_record: fullConfig.auto_record })} className="px-4 py-2 rounded-lg bg-blue-600 hover:bg-blue-700 font-semibold text-white">Save System Settings</button></div>
                 </div>
            )}
            {activeTab === 'audio' && (
                <div className="space-y-4">
                     <div><label className="block text-sm font-medium text-gray-400 mb-1">Audio Device</label><select name="audio_settings.device" value={fullConfig.audio_settings?.device} onChange={handleConfigChange} className="w-full bg-gray-700 rounded-lg px-3 py-2 appearance-none">{audioDevices.map(dev => <option key={dev.id} value={dev.id}>{dev.name} ({dev.id})</option>)}</select></div>
                     <hr className="border-gray-600"/>
                      <div><label className="block text-sm font-medium text-gray-400 mb-1">Bit Rate (Opus/Vorbis)</label>
                        <select name="audio_settings.bitrate" value={fullConfig.audio_settings?.bitrate} onChange={handleConfigChange} className="w-full bg-gray-700 rounded-lg px-3 py-2 appearance-none">
                            <option value="48000">48 kbps</option>
                            <option value="64000">64 kbps</option>
                            <option value="96000">96 kbps</option>
                            <option value="128000">128 kbps</option>
                            <option value="192000">192 kbps</option>
                            <option value="256000">256 kbps</option>
                        </select>
                         <p className="text-xs text-gray-500 mt-1">Note: Not all bitrates may be supported by your hardware. If a stream fails to start, try a different bitrate.</p>
                      </div>
                     <div><label className="block text-sm font-medium text-gray-400 mb-1">Bit Depth (Future)</label><input type="number" name="audio_settings.bit_depth" value={fullConfig.audio_settings?.bit_depth} onChange={handleConfigChange} disabled className="w-full bg-gray-900 rounded-lg px-3 py-2 cursor-not-allowed" /></div>
                      <div><label className="block text-sm font-medium text-gray-400 mb-1">Channels (Future)</label><input type="number" name="audio_settings.channels" value={fullConfig.audio_settings?.channels} onChange={handleConfigChange} disabled className="w-full bg-gray-900 rounded-lg px-3 py-2 cursor-not-allowed" /></div>
                     <div className="flex justify-end pt-4"><button onClick={() => onSave.audio(fullConfig.audio_settings)} className="px-4 py-2 rounded-lg bg-blue-600 hover:bg-blue-700 font-semibold text-white">Save Audio Settings</button></div>
                </div>
            )}
            {activeTab === 'icecast' && (
                <div className="space-y-4 max-h-[60vh] overflow-y-auto pr-2">
                    <div><label className="block text-sm font-medium text-gray-400 mb-1">Server Type</label><select name="icecast_settings.server_type" value={fullConfig.icecast_settings?.server_type} onChange={handleConfigChange} className="w-full bg-gray-700 rounded-lg px-3 py-2 appearance-none"><option value="icecast2">Icecast 2</option><option value="icecast-kh">Icecast-KH</option></select></div>
                    <hr className="border-gray-600"/>
                    <div><label className="block text-sm font-medium text-gray-400 mb-1">Server URL</label><input type="text" name="icecast_settings.url" value={fullConfig.icecast_settings?.url} onChange={handleConfigChange} className="w-full bg-gray-700 rounded-lg px-3 py-2" /></div>
                    <div><label className="block text-sm font-medium text-gray-400 mb-1">Port</label><input type="text" name="icecast_settings.port" value={fullConfig.icecast_settings?.port} onChange={handleConfigChange} className="w-full bg-gray-700 rounded-lg px-3 py-2" /></div>
                    <div><label className="block text-sm font-medium text-gray-400 mb-1">Mount Point</label><input type="text" name="icecast_settings.mount" value={fullConfig.icecast_settings?.mount} onChange={handleConfigChange} className="w-full bg-gray-700 rounded-lg px-3 py-2" /></div>
                    <div><label className="block text-sm font-medium text-gray-400 mb-1">Password</label><input type="password" name="icecast_settings.password" value={fullConfig.icecast_settings?.password} onChange={handleConfigChange} className="w-full bg-gray-700 rounded-lg px-3 py-2" /></div>
                    <hr className="border-gray-600"/>
                    <div><label className="block text-sm font-medium text-gray-400 mb-1">Stream Name</label><input type="text" name="icecast_settings.stream_name" value={fullConfig.icecast_settings?.stream_name} onChange={handleConfigChange} className="w-full bg-gray-700 rounded-lg px-3 py-2" /></div>
                    <div><label className="block text-sm font-medium text-gray-400 mb-1">Genre</label><input type="text" name="icecast_settings.stream_genre" value={fullConfig.icecast_settings?.stream_genre} onChange={handleConfigChange} className="w-full bg-gray-700 rounded-lg px-3 py-2" /></div>
                    <div><label className="block text-sm font-medium text-gray-400 mb-1">Description</label><input type="text" name="icecast_settings.stream_description" value={fullConfig.icecast_settings?.stream_description} onChange={handleConfigChange} className="w-full bg-gray-700 rounded-lg px-3 py-2" /></div>
                    <div className="flex justify-end pt-4"><button onClick={() => onSave.icecast(fullConfig.icecast_settings)} className="px-4 py-2 rounded-lg bg-blue-600 hover:bg-blue-700 font-semibold text-white">Save Icecast Settings</button></div>
                </div>
            )}
        </Modal>
    );
};

