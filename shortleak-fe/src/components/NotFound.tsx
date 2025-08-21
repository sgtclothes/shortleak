import { AlertCircle, Home, Link2 } from "lucide-react";

const NotFound = ({ shortToken, onGoHome }: { shortToken?: string; onGoHome?: () => void }) => {
    return (
        <div className="min-h-screen p-5 bg-gradient-to-br from-blue-50 via-white to-purple-50 flex items-center justify-center px-4">
            <div className="max-w-lg w-full">
                <div className="bg-white rounded-3xl shadow-2xl border border-gray-100 p-8 text-center">
                    {/* Error Icon */}
                    <div className="bg-gradient-to-r from-red-100 to-orange-100 p-4 rounded-2xl w-20 h-20 mx-auto mb-6 flex items-center justify-center">
                        <AlertCircle className="text-red-500" size={40} />
                    </div>

                    <h1 className="text-3xl font-bold text-gray-900 mb-2">Link Not Found</h1>

                    {shortToken ? (
                        <div className="space-y-4 mb-8">
                            <p className="text-gray-600">
                                The short link{" "}
                                <span className="font-mono bg-gray-100 px-2 py-1 rounded text-sm">/{shortToken}</span>{" "}
                                doesn't exist or has been removed.
                            </p>

                            <div className="bg-amber-50 border border-amber-200 rounded-xl p-4">
                                <div className="flex items-start gap-3">
                                    <AlertCircle className="text-amber-600 mt-0.5" size={20} />
                                    <div className="text-left">
                                        <p className="text-amber-800 font-medium text-sm mb-1">Possible reasons:</p>
                                        <ul className="text-amber-700 text-sm space-y-1">
                                            <li>• The link has expired</li>
                                            <li>• The link was deleted by the owner</li>
                                            <li>• You may have mistyped the URL</li>
                                        </ul>
                                    </div>
                                </div>
                            </div>
                        </div>
                    ) : (
                        <p className="text-gray-600 mb-8">The page you're looking for doesn't exist.</p>
                    )}

                    {/* 404 Illustration */}
                    <div className="text-8xl font-bold text-gray-200 mb-6 select-none">404</div>

                    {/* Action Buttons */}
                    <div className="space-y-3">
                        <button
                            onClick={onGoHome || (() => (window.location.href = "/"))}
                            className="cursor-pointer w-full bg-gradient-to-r from-blue-600 to-purple-600 text-white py-3 px-4 rounded-xl hover:from-blue-700 hover:to-purple-700 transition-all duration-200 font-medium flex items-center justify-center gap-2"
                        >
                            <Home size={18} />
                            Go to Homepage
                        </button>

                        <button
                            onClick={() => window.history.back()}
                            className="cursor-pointer w-full bg-gray-100 text-gray-700 py-3 px-4 rounded-xl hover:bg-gray-200 transition-colors font-medium"
                        >
                            Go Back
                        </button>
                    </div>

                    {/* Create New Link */}
                    <div className="mt-8 pt-6 border-t border-gray-100">
                        <p className="text-sm text-gray-500 mb-3">Want to create a short link?</p>
                        <button
                            onClick={onGoHome || (() => (window.location.href = "/"))}
                            className="cursor-pointer text-blue-600 hover:text-blue-700 font-medium text-sm flex items-center justify-center gap-1 mx-auto"
                        >
                            <Link2 size={16} />
                            Create Short Link
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default NotFound;
