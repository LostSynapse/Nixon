import React from 'react';
import { AlertTriangle, Play, Pause, Download, Edit, Lock, Unlock, Trash2 } from 'lucide-react';

const RecordingsList = ({ recordings, onPlayPause, playingFile, isAudioPlaying, onEdit, onProtect, onDelete }) => {
    if (recordings === null) {
        return <div className="bg-gray-800 p-8 rounded-lg shadow-md text-center text-gray-500"><AlertTriangle className="w-12 h-12 mx-auto mb-4" /><p>Loading recordings...</p></div>;
    }

    if (recordings.length === 0) {
        return <div className="bg-gray-800 p-8 rounded-lg shadow-md text-center text-gray-500"><AlertTriangle className="w-12 h-12 mx-auto mb-4" /><p>No recordings found.</p></div>;
    }
    
    return (
        <section>
            <h2 className="text-xl font-semibold border-b-2 border-gray-700 pb-2 mb-4">Recordings</h2>
            <div className="bg-gray-800 rounded-lg shadow-md overflow-hidden">
                <table className="w-full text-left text-sm">
                    <thead className="bg-gray-700 text-gray-300 uppercase"><tr className="hidden md:table-row"><th className="p-3 w-2/5">Name</th><th className="p-3 w-1/5">Date</th><th className="p-3 text-center">Actions</th></tr></thead>
                    <tbody className="divide-y divide-gray-700">
                        {recordings.map(rec => (
                            <tr key={rec.id} className={`flex flex-col md:table-row py-2 md:py-0 px-2 hover:bg-gray-700/50 ${playingFile === `/recordings/${rec.filename}` ? 'bg-blue-900/50' : ''}`}>
                                <td className="p-1 md:p-3 font-medium flex items-center justify-between"><span>{rec.name || rec.filename}</span><span className="md:hidden text-gray-400 text-xs">{new Date(rec.createdAt).toLocaleDateString()}</span></td>
                                <td className="p-1 md:p-3 text-gray-400 hidden md:table-cell">{new Date(rec.createdAt).toLocaleString()}</td>
                                <td className="p-1 md:p-3">
                                    <div className="flex items-center justify-center space-x-2">
                                        <button onClick={() => onPlayPause(rec.filename)} className="p-2 text-gray-400 hover:text-white">{playingFile === `/recordings/${rec.filename}` && isAudioPlaying ? <Pause size={18} /> : <Play size={18} />}</button>
                                        <a href={`/recordings/${rec.filename}`} download className="p-2 text-gray-400 hover:text-white inline-block"><Download size={18} /></a>
                                        <button onClick={() => onEdit(rec)} className="p-2 text-gray-400 hover:text-white"><Edit size={18} /></button>
                                        <button onClick={() => onProtect(rec)} className="p-2 text-gray-400 hover:text-white">{rec.protected ? <Lock size={18} className="text-yellow-400"/> : <Unlock size={18}/>}</button>
                                        <button onClick={() => !rec.protected && onDelete(rec)} disabled={rec.protected} className="p-2 text-red-500 hover:text-red-400 disabled:text-gray-600 disabled:cursor-not-allowed"><Trash2 size={18} /></button>
                                    </div>
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        </section>
    );
};

export default RecordingsList;

