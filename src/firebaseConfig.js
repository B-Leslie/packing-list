import { initializeApp } from 'firebase/app';
import {
    getAuth,
    onAuthStateChanged,
    signInAnonymously,
    signInWithCustomToken,
    createUserWithEmailAndPassword,
    signInWithEmailAndPassword,
    GoogleAuthProvider,
    signInWithPopup,
    signOut } from 'firebase/auth';
import { getFirestore, setLogLevel as setFirestoreLogLevel } from 'firebase/firestore';


// --- Firebase Configuration ---
// For local development and standard builds (like for Firebase Hosting),
// replace the placeholder values below with your actual Firebase project configuration.
// You can find these in your Firebase project settings.
const firebaseConfigValues = {
  apiKey: "AIzaSyD9uU0aiogwFEGKUHKdbCrbo7Pao8ImpqU",
  authDomain: "packing-lists-f6bdb.firebaseapp.com",
  projectId: "packing-lists-f6bdb",
  storageBucket: "packing-lists-f6bdb.firebasestorage.app",
  messagingSenderId: "486228716387",
  appId: "1:486228716387:web:c70c736cbc55243b7f4156"
};

// In a standard build environment for Firebase Hosting, we'll directly use firebaseConfigValues.
const resolvedFirebaseConfig = firebaseConfigValues;

// --- App Specific Configuration ---

// For initialAuthToken, default to null for standard anonymous sign-in flow.
// If you were using a specific custom token for initial sign-in in a special environment,
// you would handle that differently (e.g., via environment variables).
export const initialAuthToken = null;

// The 'appId' for your artifact path in Firestore.
// This defaults to using the Firebase App ID from your config.
// Ensure `firebaseConfigValues.appId` is correctly filled.
export const appId = resolvedFirebaseConfig.appId || 'default-trip-packer-app-v2-fallback';
// Added a fallback string in case resolvedFirebaseConfig.appId is somehow empty after user replacement.


// Initialize Firebase App
// Basic check to ensure the config has key values before initializing
if (!resolvedFirebaseConfig.apiKey || !resolvedFirebaseConfig.projectId) {
    console.error("Firebase configuration is missing. Please check src/firebaseConfig.js and ensure you have replaced the placeholder values with your actual Firebase project credentials.");
    // You might want to throw an error here or handle this more gracefully depending on your app's needs
}

const app = initializeApp(resolvedFirebaseConfig);

// Initialize Firebase Services
const auth = getAuth(app);
const db = getFirestore(app);

// Set Firestore log level (optional, for debugging)
// setFirestoreLogLevel('debug'); // You can uncomment this for verbose Firestore logging

// --- Helper: UUID Generator ---
export function generateUUID() {
    if (typeof crypto !== 'undefined' && crypto.randomUUID) {
        return crypto.randomUUID();
    }
    // Fallback for environments where crypto.randomUUID is not available
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
        var r = Math.random() * 16 | 0, v = c == 'x' ? r : (r & 0x3 | 0x8);
        return v.toString(16);
    });
}

export {
app,
auth,
db,
onAuthStateChanged,
signInAnonymously, // Keep if you still want a way to trigger anonymous sign-in explicitly
signInWithCustomToken, // Keep if used
createUserWithEmailAndPassword, // New
signInWithEmailAndPassword,   // New
GoogleAuthProvider,           // New
signInWithPopup,              // New
signOut                       // New
};

