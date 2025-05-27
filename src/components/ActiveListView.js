import React, { useState } from 'react';
import { generateUUID } from '../firebaseConfig';

// ActiveListView (Styled and handles categories with collapse)
function ActiveListView({ list, allLists, onUpdateList, onOpenAddListModal, onGoBack }) {
    const [newItemName, setNewItemName] = useState('');
    const [newCategoryName, setNewCategoryName] = useState('');
    const [itemForCategory, setItemForCategory] = useState({ categoryId: null, name: '' });

    if (!list) return <p className="text-center text-slate-600">List not found.</p>;

    const handleAddItem = () => {
        if (!newItemName.trim()) return;
        const newItem = { id: generateUUID(), type: 'item', name: newItemName.trim(), checked: false };
        const updatedItems = [...(list.items || []), newItem];
        onUpdateList({ ...list, items: updatedItems });
        setNewItemName('');
    };

    const handleAddCategory = () => {
        if (!newCategoryName.trim()) return;
        const newCategory = { id: generateUUID(), type: 'sublist', name: newCategoryName.trim(), items: [], checked: false, isCollapsed: false };
        const updatedItems = [...(list.items || []), newCategory];
        onUpdateList({ ...list, items: updatedItems });
        setNewCategoryName('');
    };

    const handleAddItemToCategory = (categoryId) => {
        if (!itemForCategory.name.trim() || itemForCategory.categoryId !== categoryId) return;
        const updatedListItems = (list.items || []).map(cat => {
            if (cat.id === categoryId && cat.type === 'sublist') {
                const newItem = { id: generateUUID(), name: itemForCategory.name.trim(), checked: false };
                return { ...cat, items: [...(cat.items || []), newItem] };
            }
            return cat;
        });
        onUpdateList({ ...list, items: updatedListItems });
        setItemForCategory({ categoryId: null, name: '' });
    };

    const handleToggleCheck = (itemId, categoryId = null) => {
        let updatedListItems;
        if (categoryId) {
            updatedListItems = (list.items || []).map(cat => {
                if (cat.id === categoryId && cat.type === 'sublist') {
                    const updatedSubItems = (cat.items || []).map(item =>
                        item.id === itemId ? { ...item, checked: !item.checked } : item
                    );
                    return { ...cat, items: updatedSubItems };
                }
                return cat;
            });
        } else {
            updatedListItems = (list.items || []).map(item =>
                item.id === itemId ? { ...item, checked: !item.checked } : item
            );
        }
        onUpdateList({ ...list, items: updatedListItems });
    };

    const handleDeleteItemOrCategory = (itemId, categoryId = null) => {
        let updatedListItems;
        if (categoryId) { // Delete item within a category
            updatedListItems = (list.items || []).map(cat => {
                if (cat.id === categoryId && cat.type === 'sublist') {
                    return { ...cat, items: (cat.items || []).filter(item => item.id !== itemId) };
                }
                return cat;
            });
        } else { // Delete top-level item or category
            updatedListItems = (list.items || []).filter(item => item.id !== itemId);
        }
        onUpdateList({ ...list, items: updatedListItems });
    };

    const handleToggleCategoryCollapse = (categoryId) => {
        const updatedListItems = (list.items || []).map(itemOrCat => {
            if (itemOrCat.id === categoryId && itemOrCat.type === 'sublist') {
                return { ...itemOrCat, isCollapsed: !itemOrCat.isCollapsed };
            }
            return itemOrCat;
        });
        onUpdateList({ ...list, items: updatedListItems });
    };

    // Filter out the current list from the lists available to import
    const listsToImport = allLists.filter(l => l.id !== list.id);

    return (
        <div className="bg-white p-6 sm:p-8 rounded-xl shadow-xl space-y-8">
            <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
                <h2 className="text-3xl sm:text-4xl font-bold text-slate-800">{list.name}</h2>
                <button
                    onClick={onGoBack}
                    className="bg-slate-200 hover:bg-slate-300 text-slate-700 font-semibold py-2 px-5 rounded-lg transition-colors text-sm flex items-center gap-2"
                >
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-5 h-5"><path strokeLinecap="round" strokeLinejoin="round" d="M15.75 19.5L8.25 12l7.5-7.5" /></svg>
                    All Lists
                </button>
            </div>

            {/* Add Top-Level Item/Category Forms */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div className="space-y-3 bg-slate-50 p-4 rounded-lg border border-slate-200">
                    <h3 className="text-lg font-semibold text-slate-700">Add New Item</h3>
                    <div className="flex gap-2">
                        <input type="text" value={newItemName} onKeyDown={(e) => { if (e.key === 'Enter') { e.preventDefault(); handleAddItem(); } }} lineonChange={(e) => setNewItemName(e.target.value)} placeholder="E.g., Passport" className="flex-grow p-2 border border-slate-300 rounded-md focus:ring-sky-500 focus:border-sky-500"/>
                        <button onClick={handleAddItem} className="bg-sky-500 hover:bg-sky-600 text-white font-medium py-2 px-4 rounded-md transition-colors">Add</button>
                    </div>
                </div>
                <div className="space-y-3 bg-slate-50 p-4 rounded-lg border border-slate-200">
                    <h3 className="text-lg font-semibold text-slate-700">Add New Category</h3>
                    <div className="flex gap-2">
                        <input type="text" value={newCategoryName} onKeyDown={(e) => { if (e.key === 'Enter') { e.preventDefault(); handleAddCategory(); } }} onChange={(e) => setNewCategoryName(e.target.value)} placeholder="E.g., Toiletries" className="flex-grow p-2 border border-slate-300 rounded-md focus:ring-sky-500 focus:border-sky-500"/>
                        <button onClick={handleAddCategory} className="bg-teal-500 hover:bg-teal-600 text-white font-medium py-2 px-4 rounded-md transition-colors">Add Category</button>
                    </div>
                </div>
            </div>

             {/* Add Items from Another List Button */}
            {listsToImport && listsToImport.length > 0 && (
                 <div className="pt-2">
                    <button
                        onClick={onOpenAddListModal}
                        className="w-full bg-indigo-500 hover:bg-indigo-600 text-white font-semibold py-3 px-5 rounded-lg transition-colors shadow-md hover:shadow-lg flex items-center justify-center gap-2"
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-5 h-5"><path strokeLinecap="round" strokeLinejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
                        Import Items from Another List
                    </button>
                </div>
            )}


            {/* Items & Categories List */}
            <div>
                <h3 className="text-2xl font-semibold text-slate-700 mb-4">List Contents:</h3>
                {(list.items && list.items.length > 0) ? (
                    <ul className="space-y-4">
                        {(list.items || []).map((itemOrCat) => (
                            <li key={itemOrCat.id} className={`rounded-lg ${itemOrCat.type === 'sublist' ? 'bg-slate-100 border border-slate-200' : 'bg-white border border-slate-100 shadow-sm'}`}>
                                <div className={`flex items-center justify-between p-4 ${itemOrCat.type === 'sublist' ? 'cursor-pointer hover:bg-slate-200 transition-colors' : ''}`}
                                     onClick={() => itemOrCat.type === 'sublist' && handleToggleCategoryCollapse(itemOrCat.id)}
                                >
                                    <div className="flex items-center gap-3">
                                        {itemOrCat.type === 'sublist' && (
                                            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" className={`w-5 h-5 text-slate-500 transition-transform duration-200 ${itemOrCat.isCollapsed ? 'rotate-0' : 'rotate-90'}`}>
                                                <path strokeLinecap="round" strokeLinejoin="round" d="M8.25 4.5l7.5 7.5-7.5 7.5" />
                                            </svg>
                                        )}
                                        <input
                                            type="checkbox"
                                            checked={!!itemOrCat.checked}
                                            onChange={(e) => {
                                                e.stopPropagation();
                                                handleToggleCheck(itemOrCat.id);
                                            }}
                                            className="form-checkbox h-5 w-5 text-sky-500 rounded border-slate-300 focus:ring-sky-400"
                                        />
                                        <span className={`text-lg ${itemOrCat.checked ? 'line-through text-slate-400' : 'text-slate-700'} ${itemOrCat.type === 'sublist' ? 'font-semibold' : ''}`}>{itemOrCat.name}</span>
                                    </div>
                                    <button
                                        onClick={(e) => {
                                            e.stopPropagation();
                                            handleDeleteItemOrCategory(itemOrCat.id);
                                        }}
                                        className="text-rose-500 hover:text-rose-700 p-1 rounded-full transition-colors"
                                        aria-label={`Delete ${itemOrCat.type === 'sublist' ? 'category' : 'item'}`}
                                    >
                                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-5 h-5"><path strokeLinecap="round" strokeLinejoin="round" d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12.56 0c1.153 0 2.24.03 3.22.077m3.22-.077L10.88 5.79m5.235 0c-.381.245-.754.515-1.103.804M15.002 11.177h-.01M9.925 11.177h-.01M12.004 5.79c-.355-.012-.706-.012-1.056 0M12.004 5.79L12 3"/></svg>
                                    </button>
                                </div>

                                {itemOrCat.type === 'sublist' && !itemOrCat.isCollapsed && (
                                    <div className="ml-6 mt-2 mb-4 px-4 pt-4 border-l-2 border-slate-300 space-y-3">
                                        {(itemOrCat.items && itemOrCat.items.length > 0) && itemOrCat.items.map(subItem => (
                                            <div key={subItem.id} className="flex items-center justify-between">
                                                <div className="flex items-center gap-2">
                                                    <input
                                                        type="checkbox"
                                                        checked={!!subItem.checked}
                                                        onChange={() => handleToggleCheck(subItem.id, itemOrCat.id)}
                                                        className="form-checkbox h-4 w-4 text-sky-500 rounded border-slate-300 focus:ring-sky-400"
                                                    />
                                                    <span className={`text-md ${subItem.checked ? 'line-through text-slate-400' : 'text-slate-600'}`}>{subItem.name}</span>
                                                </div>
                                                <button
                                                    onClick={() => handleDeleteItemOrCategory(subItem.id, itemOrCat.id)}
                                                    className="text-rose-400 hover:text-rose-600 p-1 rounded-full transition-colors"
                                                    aria-label="Delete sub-item"
                                                >
                                                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-4 h-4"><path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" /></svg>
                                                </button>
                                            </div>
                                        ))}
                                        <div className="flex gap-2 pt-2">
                                            <input
                                                type="text"
                                                value={itemForCategory.categoryId === itemOrCat.id ? itemForCategory.name : ''}
					    onKeyDown={(e) => { if (e.key === 'Enter') { e.preventDefault(); handleAddItemToCategory(itemOrCat.id); } }} 
                                                onChange={(e) => setItemForCategory({ categoryId: itemOrCat.id, name: e.target.value })}
                                                placeholder="Add item to this category"
                                                className="flex-grow p-2 text-sm border border-slate-300 rounded-md focus:ring-sky-500 focus:border-sky-500"
                                            />
                                            <button
                                                onClick={() => handleAddItemToCategory(itemOrCat.id)}
                                                className="bg-sky-400 hover:bg-sky-500 text-white text-sm font-medium py-2 px-3 rounded-md transition-colors"
                                            >
                                                Add
                                            </button>
                                        </div>
                                    </div>
                                )}
                            </li>
                        ))}
                    </ul>
                ) : (
                    <p className="text-slate-500 italic text-center py-4">This list is currently empty. Add some items or categories!</p>
                )}
            </div>
        </div>
    );
}

export default ActiveListView;

