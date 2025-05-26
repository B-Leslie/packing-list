import React, { useState } from 'react';

// ListSelectorView (Styled)
function ListSelectorView({ lists, onCreateList, onSelectList, onDeleteList, isLoading, setConfirmDelete }) {
    const [newListName, setNewListName] = useState('');

    const handleSubmit = (e) => {
        e.preventDefault();
        if (newListName.trim()) {
            onCreateList(newListName.trim());
            setNewListName('');
        }
    };

    return (
        <div className="space-y-8">
            <form onSubmit={handleSubmit} className="bg-white p-6 sm:p-8 rounded-xl shadow-lg">
                <h2 className="text-2xl sm:text-3xl font-semibold text-slate-700 mb-6">Create New Packing List</h2>
                <div className="flex flex-col sm:flex-row gap-4">
                    <input
                        type="text"
                        value={newListName}
                        onChange={(e) => setNewListName(e.target.value)}
                        placeholder="E.g., Weekend Camping Trip"
                        className="flex-grow p-3 border border-slate-300 rounded-lg focus:ring-2 focus:ring-sky-500 focus:border-sky-500 transition-shadow text-slate-700"
                    />
                    <button
                        type="submit"
                        className="bg-sky-500 hover:bg-sky-600 text-white font-semibold py-3 px-6 rounded-lg transition-colors shadow-md hover:shadow-lg focus:outline-none focus:ring-2 focus:ring-sky-500 focus:ring-opacity-50"
                    >
                        Create List
                    </button>
                </div>
            </form>

            <div className="bg-white p-6 sm:p-8 rounded-xl shadow-lg">
                <h2 className="text-2xl sm:text-3xl font-semibold text-slate-700 mb-6">Your Packing Lists</h2>
                {isLoading && <p className="text-slate-500">Loading lists...</p>}
                {!isLoading && lists.length === 0 && (
                    <p className="text-slate-500 italic">No packing lists yet. Create one above to get started!</p>
                )}
                {!isLoading && lists.length > 0 && (
                    <ul className="space-y-4">
                        {lists.map((list) => (
                            <li key={list.id} className="flex flex-col sm:flex-row justify-between items-center p-4 bg-slate-50 rounded-lg border border-slate-200 hover:shadow-md transition-shadow group">
                                <span className="text-lg text-slate-800 font-medium mb-2 sm:mb-0">{list.name} ({list.items?.length || 0} sections/items)</span>
                                <div className="flex gap-3">
                                    <button
                                        onClick={() => onSelectList(list.id)}
                                        className="bg-emerald-500 hover:bg-emerald-600 text-white py-2 px-4 rounded-md text-sm font-medium transition-colors shadow hover:shadow-md"
                                    >
                                        View/Edit
                                    </button>
                                    <button
                                        onClick={() => setConfirmDelete({ listId: list.id, listName: list.name })}
                                        className="bg-rose-500 hover:bg-rose-600 text-white py-2 px-4 rounded-md text-sm font-medium transition-colors shadow hover:shadow-md"
                                    >
                                        Delete
                                    </button>
                                </div>
                            </li>
                        ))}
                    </ul>
                )}
            </div>
        </div>
    );
}

export default ListSelectorView;

