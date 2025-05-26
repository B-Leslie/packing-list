import { initializeApp } from 'firebase/app';
import { getAuth, onAuthStateChanged, signInAnonymously, signInWithCustomToken } from 'firebase/auth';
import { getFirestore, setLogLevel as setFirestoreLogLevel } from 'firebase/firestore';

// --- Firebase Configuration ---
// These are expected to be globally available in the execution environment.
// For local development, you would replace these with your actual Firebase project config.
const firebaseConfig = typeof __firebase_config !== 'undefined' ? JSON.parse(__firebase_config) : {
    apiKey: "YOUR_API_KEY",
    authDomain: "YOUR_AUTH_DOMAIN",
    projectId: "YOUR_PROJECT_ID",
    storageBucket: "YOUR_STORAGE_BUCKET",
    messagingSenderId: "YOUR_MESSAGING_SENDER_ID",
    appId: "YOUR_APP_ID"
};

export const initialAuthToken = typeof __initial_auth_token !== 'undefined' ? __initial_auth_token : null;
export const appId = typeof __app_id !== 'undefined' ? __app_id : 'default-trip-packer-app-v2';

// Initialize Firebase App
const app = initializeApp(firebaseConfig);

// Initialize Firebase Services
const auth = getAuth(app);
const db = getFirestore(app);

// Set Firestore log level (optional, for debugging)
setFirestoreLogLevel('debug');

// --- Helper: UUID Generator ---
export function generateUUID() {
    return crypto.randomUUID();
}

export { app, auth, db, onAuthStateChanged, signInAnonymously, signInWithCustomToken };
