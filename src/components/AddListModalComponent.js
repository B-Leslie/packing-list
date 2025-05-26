import React, { useState, useEffect } from 'react';
import Modal from './Modal';

// AddListModalComponent (Simplified - no merge type)
function AddListModalComponent({ isOpen, onClose, allLists, onConfirmAddList }) {
    const [selectedListId, setSelectedListId] = useState('');

    useEffect(() => {
        // Pre-select the first list if available and modal opens, or if selected list is no longer valid
        if (isOpen && allLists.length > 0) {
            if (!allLists.find(l => l.id === selectedListId)) { // if current selection is invalid or empty
                 setSelectedListId(allLists[0].id);
            }
        } else if (isOpen && allLists.length === 0) {
            setSelectedListId(''); // No lists to select
        }
    }, [isOpen, allLists, selectedListId]);

    const handleConfirm = () => {
        if (selectedListId) {
            onConfirmAddList(selectedListId);
        }
    };

    if (allLists.length === 0 && isOpen) {
         return (
            <Modal isOpen={isOpen} onClose={onClose} title="Import Items From Another List">
                <p className="text-slate-600 mb-6">You don't have any other lists to import items from.</p>
                 <div className="flex justify-end">
                    <button
                        onClick={onClose}
                        className="bg-slate-500 hover:bg-slate-600 text-white font-semibold py-2 px-4 rounded-lg transition-colors"
                    >
                        Close
                    </button>
                </div>
            </Modal>
        );
    }

    return (
        <Modal isOpen={isOpen} onClose={onClose} title="Import Items From Another List">
            <div className="space-y-6">
                <div>
                    <label htmlFor="list-select" className="block text-sm font-medium text-slate-700 mb-1">
                        Select list to import as a new category:
                    </label>
                    <select
                        id="list-select"
                        value={selectedListId}
                        onChange={(e) => setSelectedListId(e.target.value)}
                        className="w-full p-3 border border-slate-300 rounded-lg focus:ring-2 focus:ring-sky-500 focus:border-sky-500 text-slate-700"
                        disabled={allLists.length === 0}
                    >
                        {allLists.map((list) => (
                            <option key={list.id} value={list.id}>
                                {list.name}
                            </option>
                        ))}
                    </select>
                </div>

                <div className="flex justify-end gap-3 pt-2">
                    <button
                        onClick={onClose}
                        className="bg-slate-200 hover:bg-slate-300 text-slate-700 font-semibold py-2 px-4 rounded-lg transition-colors"
                    >
                        Cancel
                    </button>
                    <button
                        onClick={handleConfirm}
                        disabled={!selectedListId} // Disable if no list is selected
                        className="bg-sky-500 hover:bg-sky-600 text-white font-semibold py-2 px-4 rounded-lg transition-colors shadow-md hover:shadow-lg disabled:opacity-60"
                    >
                        Import as Category
                    </button>
                </div>
            </div>
        </Modal>
    );
}

export default AddListModalComponent;

