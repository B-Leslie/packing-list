import React, { useState, useEffect, useMemo } from 'react';
import {
    db,
    auth,
    onAuthStateChanged,
    // signInAnonymously, // No longer primary, remove if not explicitly needed
    // signInWithCustomToken, // Keep if you have a specific use case for it
    initialAuthToken, // This might also be less relevant now
    appId,
    generateUUID,
    createUserWithEmailAndPassword,
    signInWithEmailAndPassword,
    GoogleAuthProvider,
    signInWithPopup,
    signOut
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
import AuthPage from './components/AuthPage'; // New Import

function App() {
    const [user, setUser] = useState(null); // Store the full Firebase user object
    const [isAuthReady, setIsAuthReady] = useState(false);
    const [authError, setAuthError] = useState(null);

    const [lists, setLists] = useState([]);
    const [currentListId, setCurrentListId] = useState(null);
    const [isLoadingLists, setIsLoadingLists] = useState(true);
    const [error, setError] = useState(null); // For general app errors
    const [showAddListModal, setShowAddListModal] = useState(false);
    const [confirmDeleteList, setConfirmDeleteList] = useState(null);

    useEffect(() => {
        const unsubscribe = onAuthStateChanged(auth, (firebaseUser) => {
            setUser(firebaseUser); // If logged out, firebaseUser will be null
            setIsAuthReady(true);
            setAuthError(null); // Clear auth errors on state change
            if (!firebaseUser) {
                // If user logs out, clear their lists and selected list from view
                setLists([]);
                setCurrentListId(null);
            }
        });
        return () => unsubscribe();
    }, []);

    const listsCollectionPath = useMemo(() => {
        if (user?.uid) { // Use optional chaining for user.uid
            return `artifacts/${appId}/users/${user.uid}/packingLists_v2`;
        }
        return null;
    }, [user?.uid]); // Depend on user.uid

    useEffect(() => {
        if (!db || !listsCollectionPath || !isAuthReady || !user) {
            if (isAuthReady && user && !listsCollectionPath) {
                 console.warn("Lists collection path is not available, but user is authenticated.");
            }
            if (!user && isAuthReady) { // If user is logged out, don't try to fetch
                setLists([]); // Clear lists
                setIsLoadingLists(false);
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
            setError("Failed to load lists.");
            setIsLoadingLists(false);
        });
        return () => unsubscribe();
    }, [db, listsCollectionPath, isAuthReady, user]); // Add user to dependency array

    // --- Authentication Handlers ---
    const handleSignUpWithEmail = async (email, password) => {
        setAuthError(null);
        try {
            await createUserWithEmailAndPassword(auth, email, password);
            // onAuthStateChanged will handle setting the user
        } catch (e) {
            console.error("Sign up error:", e);
            setAuthError(e.message);
        }
    };

    const handleSignInWithEmail = async (email, password) => {
        setAuthError(null);
        try {
            await signInWithEmailAndPassword(auth, email, password);
            // onAuthStateChanged will handle setting the user
        } catch (e) {
            console.error("Sign in error:", e);
            setAuthError(e.message);
        }
    };

    const handleSignInWithGoogle = async () => {
        setAuthError(null);
        const provider = new GoogleAuthProvider();
        try {
            await signInWithPopup(auth, provider);
            // onAuthStateChanged will handle setting the user
        } catch (e) {
            console.error("Google sign in error:", e);
            if (e.code === 'auth/popup-closed-by-user') {
                setAuthError("Sign-in process was cancelled.");
            } else {
                setAuthError(e.message);
            }
        }
    };

    const handleSignOut = async () => {
        setAuthError(null);
        try {
            await signOut(auth);
            setCurrentListId(null); // Clear current list on sign out
            // onAuthStateChanged will set user to null and clear lists
        } catch (e) {
            console.error("Sign out error:", e);
            setAuthError(e.message); // Or a generic app error
        }
    };

    // --- List CRUD Handlers (mostly unchanged, ensure they use user.uid via listsCollectionPath) ---
    const handleCreateList = async (name) => {
        if (!db || !listsCollectionPath || !name.trim() || !user) { // Check for user
            setError("You must be logged in to create a list.");
            return;
        }
        // ... (rest of the function is the same)
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
        if (!db || !listsCollectionPath || !listId || !user) return; // Check for user
        // ... (rest of the function is the same)
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
        if (!db || !listsCollectionPath || !currentListId || !updatedListObject || !user) { // Check for user
             setError("You must be logged in to update a list.");
             return;
        }
        // ... (rest of the function is the same)
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
        if (!db || !listsCollectionPath || !currentListId || !currentList || !listToAddId || !user) { // Check for user
            setError("You must be logged in to import a list.");
            return;
        }
        // ... (rest of the function is the same)
        const listToAdd = lists.find(l => l.id === listToAddId);
        if (!listToAdd) {
            setError("Selected list to import was not found.");
            return;
        }
        const newSublist = { /* ... as before ... */
            id: generateUUID(),
            type: 'sublist',
            name: listToAdd.name,
            checked: false,
            isCollapsed: false,
            items: (listToAdd.items || []).map(item => { /* ... as before ... */
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

    // --- Render Logic ---
    if (!isAuthReady) {
        return <div className="flex justify-center items-center h-screen bg-slate-100"><p className="text-xl text-slate-600 animate-pulse">Initializing Trip Packer...</p></div>;
    }

    if (!user) { // If no user is signed in, show AuthPage
        return <AuthPage
                    onSignUpWithEmail={handleSignUpWithEmail}
                    onSignInWithEmail={handleSignInWithEmail}
                    onSignInWithGoogle={handleSignInWithGoogle}
                    authError={authError}
                />;
    }

    // User is signed in, show the main app
    return (
        <div className="min-h-screen bg-slate-100 font-sans text-slate-800">
            <div className="container mx-auto p-4 sm:p-6 md:p-8">
                <header className="mb-10 text-center">
                    <div className="flex justify-between items-center mb-2">
                        <div></div> {/* Spacer */}
                        <h1 className="text-4xl sm:text-5xl font-extrabold text-transparent bg-clip-text bg-gradient-to-r from-sky-500 to-indigo-600 py-2">
                            Trip Packer Deluxe
                        </h1>
                        <button
                            onClick={handleSignOut}
                            className="bg-rose-500 hover:bg-rose-600 text-white font-semibold py-2 px-4 rounded-lg transition-colors text-sm"
                        >
                            Sign Out
                        </button>
                    </div>
                    {user.email && <p className="text-sm text-slate-500 mt-1">Signed in as: {user.email}</p>}
                    {!user.email && user.isAnonymous && <p className="text-sm text-slate-500 mt-1">Signed in anonymously</p>}
                    {!user.email && !user.isAnonymous && user.displayName && <p className="text-sm text-slate-500 mt-1">Signed in as: {user.displayName} (Google)</p>}

                </header>

                {error && ( /* For general app errors */
                    <div className="bg-rose-50 border-l-4 border-rose-500 text-rose-700 p-4 mb-6 rounded-md shadow-md" role="alert">
                        {/* ... error display ... */}
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

