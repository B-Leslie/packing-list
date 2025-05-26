import React from 'react';
import Modal from './Modal';

// Custom Confirmation Modal
function ConfirmationModal({ isOpen, onClose, title, message, onConfirm, confirmText = "Confirm", cancelText = "Cancel" }) {
    if (!isOpen) return null;

    return (
        <Modal isOpen={isOpen} onClose={onClose} title={title}>
            <p className="text-slate-600 mb-6">{message}</p>
            <div className="flex justify-end gap-3">
                <button
                    onClick={onClose}
                    className="px-4 py-2 rounded-md text-slate-700 bg-slate-100 hover:bg-slate-200 transition-colors font-medium"
                >
                    {cancelText}
                </button>
                <button
                    onClick={() => {
                        onConfirm();
                        onClose(); // Ensure modal closes after confirm
                    }}
                    className="px-4 py-2 rounded-md text-white bg-red-500 hover:bg-red-600 transition-colors font-medium"
                >
                    {confirmText}
                </button>
            </div>
        </Modal>
    );
}

export default ConfirmationModal;

