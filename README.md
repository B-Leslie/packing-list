# Trip Packer Deluxe - React App

This is a React application for creating and managing trip packing lists, built with Firebase for data storage and Tailwind CSS for styling.

## Project Structure

trip-packer-react-app/├── public/│   └── index.html            # Main HTML page├── src/│   ├── App.js                  # Main application component│   ├── firebaseConfig.js       # Firebase initialization and config│   ├── index.css               # Tailwind directives and global styles│   ├── index.js                # React entry point│   └── components/             # Reusable UI components│       ├── ActiveListView.js│       ├── AddListModalComponent.js│       ├── ConfirmationModal.js│       ├── ListSelectorView.js│       └── Modal.js├── .gitignore                  # Standard Create React App .gitignore├── package.json                # Project dependencies and scripts├── README.md                   # This file└── tailwind.config.js          # Tailwind CSS configuration
## Setup and Running

1.  **Firebase Setup:**
    * Create a Firebase project at [https://console.firebase.google.com/](https://console.firebase.google.com/).
    * Add a Web App to your Firebase project.
    * Copy your Firebase project configuration (apiKey, authDomain, etc.).
    * Replace the placeholder values in `src/firebaseConfig.js` with your actual Firebase config if you are running this locally and not in an environment where `__firebase_config` is provided.
    * In your Firebase project, enable Firestore database.
    * Set up Firestore security rules. A basic rule to allow authenticated users to read/write their own data under `artifacts/{appId}/users/{userId}/packingLists_v2/{listId}` would be:
        ```json
        rules_version = '2';
        service cloud.firestore {
          match /databases/{database}/documents {
            match /artifacts/{appId}/users/{userId}/packingLists_v2/{listDocId} {
              allow read, write: if request.auth != null && request.auth.uid == userId;
            }
          }
        }
        ```
    * Enable Anonymous sign-in in Firebase Authentication > Sign-in method.

2.  **Install Dependencies:**
    Navigate to the `trip-packer-react-app` directory in your terminal and run:
    ```bash
    npm install
    ```
    or
    ```bash
    yarn install
    ```

3.  **Tailwind CSS:**
    This project is set up to use Tailwind CSS.
    * The `package.json` includes `tailwindcss` as a dev dependency.
    * A `tailwind.config.js` is provided.
    * `src/index.css` includes the necessary Tailwind directives.
    * Ensure your `src/index.js` imports `src/index.css` as shown in the standard Create React App setup:
        ```javascript
        import React from 'react';
        import ReactDOM from 'react-dom/client';
        import './index.css'; // Import Tailwind CSS
        import App from './App';
        // ... other imports like reportWebVitals

        const root = ReactDOM.createRoot(document.getElementById('root'));
        root.render(
          <React.StrictMode>
            <App />
          </React.StrictMode>
        );
        ```

4.  **Run the Application:**
    ```bash
    npm start
    ```
    or
    ```bash
    yarn start
    ```
    This will typically open the app in your browser at `http://localhost:3000`.

## Features
* Create multiple packing lists.
* Add items and categories (sub-lists) to each list.
* Mark items/categories as packed (checked).
* Delete items, categories, or entire lists.
* Import an existing list as a new category into the current list.
* Collapsible categories for better organization.
* Data is saved to Firebase Firestore in real-time.
* Minimal and clean user interface styled with Tailwind CSS.

