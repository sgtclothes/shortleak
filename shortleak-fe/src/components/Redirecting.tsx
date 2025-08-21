import { ArrowRight, Link2, Loader2, RefreshCw } from "lucide-react";
import { useEffect, useState } from "react";

const Redirecting = ({
    shortToken,
    originalUrl,
    onRetry,
}: {
    shortToken: string;
    originalUrl: string;
    onRetry: () => void;
}) => {
    const [countdown, setCountdown] = useState(3);
    const [isRedirecting, setIsRedirecting] = useState(true);

    useEffect(() => {
        if (countdown > 0 && isRedirecting) {
            const timer = setTimeout(() => setCountdown(countdown - 1), 1000);
            return () => clearTimeout(timer);
        } else if (countdown === 0 && isRedirecting && originalUrl) {
            window.location.href = originalUrl;
        }
    }, [countdown, isRedirecting, originalUrl]);

    const handleRedirectNow = () => {
        if (originalUrl) {
            window.location.href = originalUrl;
        }
    };

    const handleCancel = () => {
        setIsRedirecting(false);
    };

    return (
        <div className="min-h-screen bg-gradient-to-br from-blue-50 via-white to-purple-50 flex items-center justify-center px-4">
            <div className="max-w-md w-full">
                <div className="bg-white rounded-3xl shadow-2xl border border-gray-100 p-8 text-center">
                    {/* Logo */}
                    <div className="bg-gradient-to-r from-blue-600 to-purple-600 p-3 rounded-2xl w-16 h-16 mx-auto mb-6 flex items-center justify-center">
                        <Link2 className="text-white" size={32} />
                    </div>

                    {originalUrl ? (
                        <>
                            <h1 className="text-2xl font-bold text-gray-900 mb-2">Redirecting...</h1>
                            <p className="text-gray-600 mb-6">
                                You will be redirected to your destination in{" "}
                                <span className="font-bold text-blue-600">{countdown}</span> seconds
                            </p>

                            {/* Animated Progress Ring */}
                            <div className="relative w-20 h-20 mx-auto mb-6">
                                <svg className="w-20 h-20 transform -rotate-90" viewBox="0 0 80 80">
                                    <circle
                                        cx="40"
                                        cy="40"
                                        r="35"
                                        stroke="currentColor"
                                        strokeWidth="6"
                                        fill="transparent"
                                        className="text-gray-200"
                                    />
                                    <circle
                                        cx="40"
                                        cy="40"
                                        r="35"
                                        stroke="currentColor"
                                        strokeWidth="6"
                                        fill="transparent"
                                        strokeDasharray={`${2 * Math.PI * 35}`}
                                        strokeDashoffset={`${2 * Math.PI * 35 * (countdown / 3)}`}
                                        className="text-blue-600 transition-all duration-1000 ease-linear"
                                        strokeLinecap="round"
                                    />
                                </svg>
                                <div className="absolute inset-0 flex items-center justify-center">
                                    <span className="text-2xl font-bold text-blue-600">{countdown}</span>
                                </div>
                            </div>

                            {/* Destination URL */}
                            <div className="bg-gray-50 rounded-xl p-4 mb-6">
                                <p className="text-sm text-gray-500 mb-1">Destination:</p>
                                <p className="text-sm font-medium text-gray-800 truncate" title={originalUrl}>
                                    {originalUrl}
                                </p>
                            </div>

                            {/* Action Buttons */}
                            <div className="space-y-3">
                                {isRedirecting ? (
                                    <>
                                        <button
                                            onClick={handleRedirectNow}
                                            className="w-full bg-gradient-to-r from-blue-600 to-purple-600 text-white py-3 px-4 rounded-xl hover:from-blue-700 hover:to-purple-700 transition-all duration-200 font-medium flex items-center justify-center gap-2"
                                        >
                                            Redirect Now <ArrowRight size={18} />
                                        </button>
                                        <button
                                            onClick={handleCancel}
                                            className="w-full bg-gray-100 text-gray-700 py-3 px-4 rounded-xl hover:bg-gray-200 transition-colors font-medium"
                                        >
                                            Cancel
                                        </button>
                                    </>
                                ) : (
                                    <button
                                        onClick={handleRedirectNow}
                                        className="w-full bg-gradient-to-r from-blue-600 to-purple-600 text-white py-3 px-4 rounded-xl hover:from-blue-700 hover:to-purple-700 transition-all duration-200 font-medium flex items-center justify-center gap-2"
                                    >
                                        Go to Destination <ArrowRight size={18} />
                                    </button>
                                )}
                            </div>
                        </>
                    ) : (
                        <>
                            <div className="animate-spin w-12 h-12 mx-auto mb-4">
                                <Loader2 className="text-blue-600" size={48} />
                            </div>
                            <h1 className="text-2xl font-bold text-gray-900 mb-2">Loading...</h1>
                            <p className="text-gray-600 mb-6">Fetching destination for /{shortToken}</p>
                            {onRetry && (
                                <button
                                    onClick={onRetry}
                                    className="bg-gradient-to-r from-blue-600 to-purple-600 text-white py-2 px-6 rounded-xl hover:from-blue-700 hover:to-purple-700 transition-all duration-200 font-medium flex items-center justify-center gap-2 mx-auto"
                                >
                                    <RefreshCw size={16} />
                                    Retry
                                </button>
                            )}
                        </>
                    )}

                    {/* Short URL Display */}
                    <div className="mt-6 pt-6 border-t border-gray-100">
                        <p className="text-xs text-gray-400 mb-1">Short URL</p>
                        <div className="flex items-center justify-center gap-2 text-sm font-mono">
                            <span className="text-gray-600">shortleak.com/</span>
                            <span className="text-blue-600 font-semibold">{shortToken}</span>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default Redirecting;
