import React, { useState, useEffect, useMemo } from 'react';
import {
    db,
    auth,
    onAuthStateChanged,
    signInAnonymously,
    signInWithCustomToken,
    initialAuthToken,
    appId,
    generateUUID
} from './firebaseConfig';
import {
    collection,
    addDoc,
    doc,
    deleteDoc,
    updateDoc,
    onSnapshot,
    serverTimestamp
} from 'firebase/firestore';

import ListSelectorView from './components/ListSelectorView';
import ActiveListView from './components/ActiveListView';
import AddListModalComponent from './components/AddListModalComponent';
import ConfirmationModal from './components/ConfirmationModal';


// --- Main App Component ---
function App() {
    const [userId, setUserId] = useState(null);
    const [isAuthReady, setIsAuthReady] = useState(false);

    const [lists, setLists] = useState([]);
    const [currentListId, setCurrentListId] = useState(null);
    const [isLoadingLists, setIsLoadingLists] = useState(true);
    const [error, setError] = useState(null);
    const [showAddListModal, setShowAddListModal] = useState(false);
    const [confirmDeleteList, setConfirmDeleteList] = useState(null);

    useEffect(() => {
        const unsubscribe = onAuthStateChanged(auth, async (user) => {
            if (user) {
                setUserId(user.uid);
            } else {
                try {
                    if (initialAuthToken) {
                        await signInWithCustomToken(auth, initialAuthToken);
                    } else {
                        await signInAnonymously(auth);
                    }
                } catch (e) {
                    console.error("Authentication error during initial sign-in:", e);
                    await signInAnonymously(auth);
                }
            }
            setIsAuthReady(true);
        });
        return () => unsubscribe();
    }, []);


    const listsCollectionPath = useMemo(() => {
        if (userId) {
            return `artifacts/${appId}/users/${userId}/packingLists_v2`;
        }
        return null;
    }, [userId]);

    useEffect(() => {
        if (!db || !listsCollectionPath || !isAuthReady) {
            if(isAuthReady && !listsCollectionPath && userId) {
                 console.warn("Lists collection path is not available, cannot fetch lists yet.");
            }
            return;
        }

        setIsLoadingLists(true);
        const q = collection(db, listsCollectionPath);
        const unsubscribe = onSnapshot(q, (snapshot) => {
            const fetchedLists = snapshot.docs.map(doc => ({ id: doc.id, ...doc.data() }));
            setLists(fetchedLists);
            setIsLoadingLists(false);
            setError(null);
        }, (err) => {
            console.error("Error fetching lists:", err);
            setError("Failed to load lists. Please check your connection or try again later.");
            setIsLoadingLists(false);
        });

        return () => unsubscribe();
    }, [db, listsCollectionPath, isAuthReady, userId]);

    const handleCreateList = async (name) => {
        if (!db || !listsCollectionPath || !name.trim()) {
            setError("Cannot create list: missing database connection, user context, or list name.");
            return;
        }
        try {
            const newListRef = await addDoc(collection(db, listsCollectionPath), {
                name: name.trim(),
                items: [],
                createdAt: serverTimestamp()
            });
            setCurrentListId(newListRef.id);
            setError(null);
        } catch (e) {
            console.error("Error creating list:", e);
            setError("Failed to create the new list.");
        }
    };

    const handleDeleteList = async (listId) => {
        if (!db || !listsCollectionPath || !listId) return;

        if (currentListId === listId) {
            setCurrentListId(null);
        }
        try {
            await deleteDoc(doc(db, listsCollectionPath, listId));
            setError(null);
        } catch (e) {
            console.error("Error deleting list:", e);
            setError("Failed to delete the list.");
        }
        setConfirmDeleteList(null);
    };

    const handleSelectList = (listId) => {
        setCurrentListId(listId);
        setError(null);
    };

    const currentList = useMemo(() => {
        return lists.find(l => l.id === currentListId);
    }, [lists, currentListId]);

    const handleUpdateCurrentList = async (updatedListObject) => {
        if (!db || !listsCollectionPath || !currentListId || !updatedListObject) {
             setError("Cannot update list: essential context or data is missing.");
             return;
        }
        try {
            const { id, createdAt, ...updatePayload } = updatedListObject;
            await updateDoc(doc(db, listsCollectionPath, currentListId), updatePayload);
            setError(null);
        } catch (e) {
            console.error("Error updating list:", e);
            setError("Failed to update the list details.");
        }
    };


    const handleConfirmAddExistingList = async (listToAddId) => {
        if (!db || !listsCollectionPath || !currentListId || !currentList || !listToAddId) {
            setError("Cannot import list: context or selection missing.");
            return;
        }

        const listToAdd = lists.find(l => l.id === listToAddId);
        if (!listToAdd) {
            setError("Selected list to import was not found.");
            return;
        }

        const newSublist = {
            id: generateUUID(),
            type: 'sublist',
            name: listToAdd.name,
            checked: false,
            isCollapsed: false,
            items: (listToAdd.items || []).map(item => {
                const newItem = {
                    id: generateUUID(),
                    name: item.name || 'Unnamed Item',
                    checked: item.checked || false,
                    type: item.type === 'sublist' ? 'sublist' : 'item',
                };
                if (newItem.type === 'sublist') {
                    newItem.isCollapsed = item.isCollapsed === undefined ? false : item.isCollapsed;
                    newItem.items = (item.items || []).map(subItem => ({
                        id: generateUUID(),
                        name: subItem.name || 'Unnamed Sub-Item',
                        checked: subItem.checked || false,
                    }));
                }
                return newItem;
            }).filter(item => item.name)
        };

        const updatedItems = [...(currentList.items || []), newSublist];
        handleUpdateCurrentList({ ...currentList, items: updatedItems });
        setShowAddListModal(false);
        setError(null);
    };

    if (!isAuthReady) {
        return <div className="flex justify-center items-center h-screen bg-slate-100"><p className="text-xl text-slate-600 animate-pulse">Initializing Trip Packer...</p></div>;
    }
     if (!db && isAuthReady) {
        return (
            <div className="flex flex-col justify-center items-center h-screen bg-slate-100 p-4 text-center">
                <h1 className="text-2xl font-bold text-rose-600 mb-4">Application Initialization Failed</h1>
                <p className="text-rose-500">{error || "Could not connect to the database. Please ensure Firebase is configured correctly and check the console for more details."}</p>
            </div>
        );
    }


    return (
        <div className="min-h-screen bg-slate-100 font-sans text-slate-800">
            <div className="container mx-auto p-4 sm:p-6 md:p-8">
                <header className="mb-10 text-center">
                    <h1 className="text-4xl sm:text-5xl font-extrabold text-transparent bg-clip-text bg-gradient-to-r from-sky-500 to-indigo-600 py-2">
                        Trip Packer Deluxe
                    </h1>
                    {userId && <p className="text-xs text-slate-400 mt-1">User ID: {userId}</p>}
                </header>

                {error && (
                    <div className="bg-rose-50 border-l-4 border-rose-500 text-rose-700 p-4 mb-6 rounded-md shadow-md" role="alert">
                        <div className="flex">
                            <div className="py-1"><svg className="fill-current h-6 w-6 text-rose-500 mr-4" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20"><path d="M2.93 17.07A10 10 0 1 1 17.07 2.93 10 10 0 0 1 2.93 17.07zM11.414 10l2.829-2.828a1 1 0 1 0-1.414-1.414L10 8.586 7.172 5.757a1 1 0 0 0-1.414 1.414L8.586 10l-2.829 2.828a1 1 0 1 0 1.414 1.414L10 11.414l2.829 2.829a1 1 0 0 0 1.414-1.414L11.414 10z"/></svg></div>
                            <div><p className="font-bold">Error</p><p className="text-sm">{error}</p></div>
                            <button onClick={() => setError(null)} className="ml-auto text-rose-500 hover:text-rose-700 p-1"><svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-5 h-5"><path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" /></svg></button>
                        </div>
                    </div>
                )}

                {!currentListId ? (
                    <ListSelectorView
                        lists={lists}
                        onCreateList={handleCreateList}
                        onSelectList={handleSelectList}
                        isLoading={isLoadingLists}
                        setConfirmDelete={setConfirmDeleteList}
                    />
                ) : (
                    currentList && (
                        <ActiveListView
                            list={currentList}
                            allLists={lists}
                            onUpdateList={handleUpdateCurrentList}
                            onOpenAddListModal={() => setShowAddListModal(true)}
                            onGoBack={() => {setCurrentListId(null); setError(null);}}
                        />
                    )
                )}

                <AddListModalComponent
                    isOpen={showAddListModal}
                    onClose={() => setShowAddListModal(false)}
                    allLists={lists.filter(l => l.id !== currentListId)}
                    onConfirmAddList={handleConfirmAddExistingList}
                />

                <ConfirmationModal
                    isOpen={!!confirmDeleteList}
                    onClose={() => setConfirmDeleteList(null)}
                    title="Confirm Deletion"
                    message={`Are you sure you want to delete the list "${confirmDeleteList?.listName}"? This action cannot be undone.`}
                    onConfirm={() => confirmDeleteList && handleDeleteList(confirmDeleteList.listId)}
                    confirmText="Delete List"
                />
            </div>
        </div>
    );
}

export default App;

