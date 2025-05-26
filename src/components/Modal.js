import React from 'react';

// Modal Component (Styled)
function Modal({ isOpen, onClose, title, children }) {
    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 bg-slate-900 bg-opacity-50 backdrop-blur-sm flex justify-center items-center p-4 z-50 transition-opacity duration-300 ease-in-out">
            <div className="bg-white p-6 rounded-xl shadow-2xl w-full max-w-lg transform transition-all duration-300 ease-in-out scale-95 opacity-0 animate-modalShow">
                <div className="flex justify-between items-center mb-6">
                    <h3 className="text-2xl font-semibold text-slate-800">{title}</h3>
                    <button
                        onClick={onClose}
                        className="text-slate-400 hover:text-slate-600 p-1 rounded-full transition-colors"
                        aria-label="Close modal"
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor" className="w-7 h-7">
                            <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
                        </svg>
                    </button>
                </div>
                {children}
            </div>
            {/* Animation for modal appearance */}
            <style jsx global>{`
                @keyframes modalShow {
                    to {
                        opacity: 1;
                        transform: scale(1);
                    }
                }
                .animate-modalShow {
                    animation: modalShow 0.3s forwards;
                }
            `}</style>
        </div>
    );
}

export default Modal;

