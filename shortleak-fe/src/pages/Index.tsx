/* eslint-disable @typescript-eslint/no-explicit-any */
import React, { useState, useEffect } from "react";
import { create } from "zustand";
import { persist } from "zustand/middleware";
import {
    Link2,
    Copy,
    ExternalLink,
    BarChart3,
    User,
    LogOut,
    X,
    Eye,
    TrendingUp,
    Calendar,
    Menu,
    Trash,
} from "lucide-react";
import { ToastContainer, toast } from "react-toastify";

const baseUrlAPI = import.meta.env.VITE_BASE_URL_API ?? "http://173.249.46.3:8090";
const baseUrl = import.meta.env.VITE_BASE_URL ?? "http://173.249.46.3:5173";
// Types
interface Link {
    id: string;
    short_token: string;
    url: string;
    CreatedAt: string;
}

interface LinkStats {
    uniqueVisitors: number;
    totalVisits: number;
    link: {
        CreatedAt: string;
        url: string;
        short_token: string;
    };
}

interface User {
    id: string;
    email: string;
    name: string;
}

interface AuthStore {
    user: User | null;
    token: string | null;
    isAuthenticated: boolean;
    login: (email: string, password: string) => Promise<boolean>;
    register: (fullname: string, email: string, password: string) => Promise<boolean>;
    logout: () => void;
    setUser: (user: User, token: string) => void;
}

interface LinkStore {
    links: Link[];
    isLoading: boolean;
    fetchLinks: () => Promise<void>;
    createLink: (url: string) => Promise<string | null>;
    deleteLink: (shortToken: string) => Promise<void>;
    setLinks: (links: Link[]) => void;
}

// Zustand stores
const useAuthStore = create<AuthStore>()(
    persist(
        (set, get) => ({
            user: null,
            token: null,
            isAuthenticated: false,
            login: async (email: string, password: string) => {
                try {
                    const response = await fetch(`${baseUrlAPI}/api/auth/login`, {
                        method: "POST",
                        headers: { "Content-Type": "application/json" },
                        credentials: "include",
                        body: JSON.stringify({ email, password }),
                    });

                    if (response.ok) {
                        const data = await response.json();
                        toast.success(data.message);
                        set({
                            user: data.user,
                            token: data.token,
                            isAuthenticated: true,
                        });
                        return true;
                    } else {
                        const errorData = await response.json();
                        toast.error(errorData.error || "Failed to login");
                    }
                    return false;
                } catch (error) {
                    console.error("Login error:", error);
                    return false;
                }
            },
            register: async (fullname: string, email: string, password: string) => {
                try {
                    const response = await fetch(`${baseUrlAPI}/api/auth/register`, {
                        method: "POST",
                        headers: { "Content-Type": "application/json" },
                        body: JSON.stringify({ fullname, email, password }),
                    });

                    if (response.ok) {
                        const data = await response.json();
                        toast.success(data.message);
                        await useAuthStore.getState().login(email, password);
                        set({
                            user: data.user,
                            token: data.token,
                            isAuthenticated: true,
                        });
                        return true;
                    } else {
                        const errorData = await response.json();
                        toast.error(errorData.error || "Failed to register");
                    }
                    return false;
                } catch (error) {
                    console.error("Register error:", error);
                    return false;
                }
            },
            logout: async () => {
                try {
                    const { token } = get();
                    if (token) {
                        await fetch(`${baseUrlAPI}/api/auth/logout`, {
                            method: "POST",
                            headers: {
                                Authorization: `Bearer ${token}`,
                                "Content-Type": "application/json",
                            },
                        });
                        toast.success("Logout successful");
                    }
                } catch (error) {
                    console.error("Logout error:", error);
                } finally {
                    set({ user: null, token: null, isAuthenticated: false });
                }
            },
            setUser: (user: User, token: string) => {
                set({ user, token, isAuthenticated: true });
            },
        }),
        {
            name: "auth-storage",
        }
    )
);

const useLinkStore = create<LinkStore>()((set, get) => ({
    links: [],
    isLoading: false,
    fetchLinks: async () => {
        const { token } = useAuthStore.getState();
        if (!token) return;

        set({ isLoading: true });
        try {
            const response = await fetch(`${baseUrlAPI}/api/links/user`, {
                headers: { Authorization: `Bearer ${token}` },
                credentials: "include",
            });
            if (response.ok) {
                const data = await response.json();
                set({ links: data || [] });
            } else {
                const errorData = await response.json();
                toast.error(errorData.error || "Failed to fetch links");
            }
        } catch (error) {
            console.error("Fetch links error:", error);
        } finally {
            set({ isLoading: false });
        }
    },
    deleteLink: async (shortToken: string) => {
        const { token } = useAuthStore.getState();
        if (!token) return;

        try {
            const response = await fetch(`${baseUrlAPI}/api/links/${shortToken}`, {
                method: "DELETE",
                headers: { Authorization: `Bearer ${token}` },
                credentials: "include",
            });
            if (response.ok) {
                toast.success("Link deleted successfully");
                get().fetchLinks();
            } else {
                const errorData = await response.json();
                toast.error(errorData.error || "Failed to delete link");
            }
        } catch (error) {
            console.error("Delete link error:", error);
        }
    },
    createLink: async (url: string) => {
        const { token } = useAuthStore.getState();
        if (!token) return null;

        try {
            const response = await fetch(`${baseUrlAPI}/shorten`, {
                method: "POST",
                headers: {
                    Authorization: `Bearer ${token}`,
                    "Content-Type": "application/json",
                },
                credentials: "include",
                body: JSON.stringify({ url }),
            });

            if (response.ok) {
                const data = await response.json();
                // Refresh links after creating
                get().fetchLinks();
                return data.shortToken;
            } else {
                const errorData = await response.json();
                toast.error(errorData.error || "Failed to create link");
            }
            return null;
        } catch (error: any) {
            console.error("Create link error:", error);
            return null;
        }
    },
    setLinks: (links: Link[]) => set({ links }),
}));

// Components
const Modal: React.FC<{ isOpen: boolean; onClose: () => void; children: React.ReactNode }> = ({
    isOpen,
    onClose,
    children,
}) => {
    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50 backdrop-blur-sm">
            <div className="relative w-full max-w-md mx-4">
                <div className="bg-white rounded-2xl shadow-2xl border border-gray-100">
                    <button
                        onClick={onClose}
                        className="absolute right-4 top-4 p-2 rounded-full hover:bg-gray-100 transition-colors"
                    >
                        <X size={20} className="text-gray-500 cursor-pointer" />
                    </button>
                    {children}
                </div>
            </div>
        </div>
    );
};

const LoginModal: React.FC<{ isOpen: boolean; onClose: () => void; onSwitchToRegister: () => void }> = ({
    isOpen,
    onClose,
    onSwitchToRegister,
}) => {
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState("");
    const login = useAuthStore((state) => state.login);

    const handleSubmit = async () => {
        setIsLoading(true);
        setError("");

        const success = await login(email, password);
        if (success) {
            onClose();
            setEmail("");
            setPassword("");
        } else {
            setError("Invalid email or password");
        }
        setIsLoading(false);
    };

    return (
        <Modal isOpen={isOpen} onClose={onClose}>
            <div className="p-8">
                <h2 className="text-2xl font-bold text-gray-900 mb-2">Welcome back</h2>
                <p className="text-gray-600 mb-6">Sign in to your account</p>

                <div className="space-y-4">
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">Email</label>
                        <input
                            type="email"
                            value={email}
                            onChange={(e) => setEmail(e.target.value)}
                            className="w-full px-4 py-3 border border-gray-300 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
                            placeholder="Enter your email"
                            required
                        />
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">Password</label>
                        <input
                            type="password"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            className="w-full px-4 py-3 border border-gray-300 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
                            placeholder="Enter your password"
                            required
                        />
                    </div>

                    {error && <p className="text-red-500 text-sm">{error}</p>}

                    <button
                        onClick={handleSubmit}
                        disabled={isLoading}
                        className="cursor-pointer w-full bg-gradient-to-r from-blue-600 to-purple-600 text-white py-3 px-4 rounded-xl hover:from-blue-700 hover:to-purple-700 transition-all duration-200 font-medium disabled:opacity-50"
                    >
                        {isLoading ? "Signing in..." : "Sign in"}
                    </button>
                </div>

                <p className="text-center text-gray-600 mt-6">
                    Don't have an account?{" "}
                    <button
                        onClick={onSwitchToRegister}
                        className="cursor-pointer text-blue-600 hover:text-blue-700 font-medium"
                    >
                        Sign up
                    </button>
                </p>
            </div>
        </Modal>
    );
};

const RegisterModal: React.FC<{ isOpen: boolean; onClose: () => void; onSwitchToLogin: () => void }> = ({
    isOpen,
    onClose,
    onSwitchToLogin,
}) => {
    const [fullname, setFullname] = useState("");
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState("");
    const register = useAuthStore((state) => state.register);

    const handleSubmit = async () => {
        setIsLoading(true);
        setError("");

        const success = await register(fullname, email, password);
        if (success) {
            onClose();
            setFullname("");
            setEmail("");
            setPassword("");
        } else {
            setError("Registration failed");
        }
        setIsLoading(false);
    };

    return (
        <Modal isOpen={isOpen} onClose={onClose}>
            <div className="p-8">
                <h2 className="text-2xl font-bold text-gray-900 mb-2">Create account</h2>
                <p className="text-gray-600 mb-6">Get started with Shortleak</p>

                <div className="space-y-4">
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">Fullname</label>
                        <input
                            type="text"
                            value={fullname}
                            onChange={(e) => setFullname(e.target.value)}
                            className="w-full px-4 py-3 border border-gray-300 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
                            placeholder="Enter your name"
                            required
                        />
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">Email</label>
                        <input
                            type="email"
                            value={email}
                            onChange={(e) => setEmail(e.target.value)}
                            className="w-full px-4 py-3 border border-gray-300 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
                            placeholder="Enter your email"
                            required
                        />
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">Password</label>
                        <input
                            type="password"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            className="w-full px-4 py-3 border border-gray-300 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
                            placeholder="Create a password"
                            required
                        />
                    </div>

                    {error && <p className="text-red-500 text-sm">{error}</p>}

                    <button
                        onClick={handleSubmit}
                        disabled={isLoading}
                        className="cursor-pointer w-full bg-gradient-to-r from-blue-600 to-purple-600 text-white py-3 px-4 rounded-xl hover:from-blue-700 hover:to-purple-700 transition-all duration-200 font-medium disabled:opacity-50"
                    >
                        {isLoading ? "Creating account..." : "Create account"}
                    </button>
                </div>

                <p className="text-center text-gray-600 mt-6">
                    Already have an account?{" "}
                    <button
                        onClick={onSwitchToLogin}
                        className="text-blue-600 hover:text-blue-700 font-medium cursor-pointer"
                    >
                        Sign in
                    </button>
                </p>
            </div>
        </Modal>
    );
};

const StatsModal: React.FC<{
    isOpen: boolean;
    onClose: () => void;
    shortToken: string;
}> = ({ isOpen, onClose, shortToken }) => {
    const [stats, setStats] = useState<LinkStats | null>(null);
    const [isLoading, setIsLoading] = useState(false);
    const { token } = useAuthStore();

    useEffect(() => {
        if (isOpen && shortToken) {
            fetchStats();
        }
    }, [isOpen, shortToken]);

    const fetchStats = async () => {
        if (!token) return;

        setIsLoading(true);
        try {
            const response = await fetch(`${baseUrlAPI}/stats/${shortToken}`, {
                headers: { Authorization: `Bearer ${token}` },
                credentials: "include",
            });

            if (response.ok) {
                const data = await response.json();
                console.log("Fetched stats:", data);
                setStats(data);
            }
        } catch (error) {
            console.error("Fetch stats error:", error);
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <Modal isOpen={isOpen} onClose={onClose}>
            <div className="p-8">
                <h2 className="text-2xl font-bold text-gray-900 mb-6 flex items-center gap-2">
                    <BarChart3 size={24} className="text-blue-600" />
                    Link Statistics
                </h2>

                {isLoading ? (
                    <div className="flex items-center justify-center py-8">
                        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
                    </div>
                ) : stats ? (
                    <div className="space-y-6">
                        <div className="bg-gray-50 rounded-xl p-4">
                            <p className="text-sm text-gray-600 mb-1">Original URL</p>
                            <p className="font-medium text-gray-900 break-all">{stats.link.url}</p>
                        </div>

                        <div className="grid grid-cols-2 gap-4">
                            <div className="bg-blue-50 rounded-xl p-4">
                                <div className="flex items-center gap-2 mb-2">
                                    <Eye className="text-blue-600" size={20} />
                                    <span className="text-sm font-medium text-blue-600">Unique Visitors</span>
                                </div>
                                <p className="text-2xl font-bold text-blue-700">{stats.uniqueVisitors}</p>
                            </div>

                            <div className="bg-green-50 rounded-xl p-4">
                                <div className="flex items-center gap-2 mb-2">
                                    <TrendingUp className="text-green-600" size={20} />
                                    <span className="text-sm font-medium text-green-600">Total Visits</span>
                                </div>
                                <p className="text-2xl font-bold text-green-700">{stats.totalVisits}</p>
                            </div>
                        </div>

                        <div className="bg-purple-50 rounded-xl p-4">
                            <div className="flex items-center gap-2 mb-2">
                                <Calendar className="text-purple-600" size={20} />
                                <span className="text-sm font-medium text-purple-600">Created</span>
                            </div>
                            <p className="text-purple-700">
                                {new Date(stats.link.CreatedAt).toLocaleDateString("en-US", {
                                    year: "numeric",
                                    month: "long",
                                    day: "numeric",
                                    hour: "2-digit",
                                    minute: "2-digit",
                                })}
                            </p>
                        </div>
                    </div>
                ) : (
                    <p className="text-gray-500 text-center py-8">Unable to load statistics</p>
                )}
            </div>
        </Modal>
    );
};

const Index: React.FC = () => {
    const [url, setUrl] = useState("");
    const [shortUrl, setShortUrl] = useState("");
    const [isLoading, setIsLoading] = useState(false);
    const [showLoginModal, setShowLoginModal] = useState(false);
    const [showRegisterModal, setShowRegisterModal] = useState(false);
    const [showStatsModal, setShowStatsModal] = useState(false);
    const [selectedLinkToken, setSelectedLinkToken] = useState("");
    const [copied, setCopied] = useState(false);

    const { user, isAuthenticated, logout } = useAuthStore();
    const { links, fetchLinks, createLink, deleteLink } = useLinkStore();
    const [menuOpen, setMenuOpen] = useState(false);

    useEffect(() => {
        if (isAuthenticated) {
            fetchLinks();
        }
    }, [isAuthenticated]);

    const handleCreateLink = async () => {
        if (!isAuthenticated) {
            setShowLoginModal(true);
            return;
        }

        setIsLoading(true);
        const result = await createLink(url);
        if (result) {
            setShortUrl(result);
            setUrl("");
        }
        setIsLoading(false);
    };

    const handleDeleteLink = async (shortToken: string) => {
        if (!isAuthenticated) {
            setShowLoginModal(true);
            return;
        }

        setIsLoading(true);
        await deleteLink(shortToken);
        setIsLoading(false);
    };

    const copyToClipboard = async (text: string) => {
        await navigator.clipboard.writeText(text);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };

    const handleShowStats = (shortToken: string) => {
        setSelectedLinkToken(shortToken);
        setShowStatsModal(true);
    };

    return (
        <div className="min-h-screen bg-gradient-to-br from-blue-50 via-white to-purple-50">
            {/* Navigation */}
            <nav className="bg-white/70 backdrop-blur-md border-b border-gray-200 sticky top-0 z-40">
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                    <div className="flex justify-between items-center h-16">
                        {/* Logo */}
                        <div className="flex items-center gap-3">
                            <div className="bg-gradient-to-r from-blue-600 to-purple-600 p-2 rounded-xl">
                                <Link2 className="text-white" size={24} />
                            </div>
                            <span className="text-2xl font-bold bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
                                Shortleak
                            </span>
                        </div>

                        {/* Desktop Menu */}
                        <div className="hidden md:flex items-center gap-4">
                            {isAuthenticated ? (
                                <>
                                    <div className="flex items-center gap-2 text-gray-700">
                                        <User size={20} />
                                        <span className="font-medium">{user?.email}</span>
                                    </div>
                                    <button
                                        onClick={logout}
                                        className="cursor-pointer flex items-center gap-2 px-4 py-2 text-red-600 hover:bg-red-50 rounded-xl transition-colors"
                                    >
                                        <LogOut size={18} />
                                        Logout
                                    </button>
                                </>
                            ) : (
                                <>
                                    <button
                                        onClick={() => setShowLoginModal(true)}
                                        className="cursor-pointer px-4 py-2 text-gray-700 hover:bg-gray-100 rounded-xl transition-colors"
                                    >
                                        Sign in
                                    </button>
                                    <button
                                        onClick={() => setShowRegisterModal(true)}
                                        className="cursor-pointer px-4 py-2 bg-gradient-to-r from-blue-600 to-purple-600 text-white rounded-xl hover:from-blue-700 hover:to-purple-700 transition-all"
                                    >
                                        Sign up
                                    </button>
                                </>
                            )}
                        </div>

                        {/* Mobile Hamburger */}
                        <div className="md:hidden">
                            <button
                                onClick={() => setMenuOpen(!menuOpen)}
                                className="p-2 rounded-md text-gray-600 hover:bg-gray-100 transition-colors"
                            >
                                {menuOpen ? <X size={24} /> : <Menu size={24} />}
                            </button>
                        </div>
                    </div>
                </div>

                {/* Mobile Menu */}
                {menuOpen && (
                    <div className="md:hidden border-t border-gray-200 bg-white/95 backdrop-blur-md">
                        <div className="px-4 py-4 space-y-3">
                            {isAuthenticated ? (
                                <>
                                    <div className="flex items-center gap-2 text-gray-700">
                                        <User size={20} />
                                        <span className="font-medium">{user?.email}</span>
                                    </div>
                                    <button
                                        onClick={logout}
                                        className="w-full text-left flex items-center gap-2 px-4 py-2 text-red-600 hover:bg-red-50 rounded-xl transition-colors"
                                    >
                                        <LogOut size={18} />
                                        Logout
                                    </button>
                                </>
                            ) : (
                                <>
                                    <button
                                        onClick={() => {
                                            setShowLoginModal(true);
                                            setMenuOpen(false);
                                        }}
                                        className="w-full text-left px-4 py-2 text-gray-700 hover:bg-gray-100 rounded-xl transition-colors"
                                    >
                                        Sign in
                                    </button>
                                    <button
                                        onClick={() => {
                                            setShowRegisterModal(true);
                                            setMenuOpen(false);
                                        }}
                                        className="w-full text-left px-4 py-2 bg-gradient-to-r from-blue-600 to-purple-600 text-white rounded-xl hover:from-blue-700 hover:to-purple-700 transition-all"
                                    >
                                        Sign up
                                    </button>
                                </>
                            )}
                        </div>
                    </div>
                )}
            </nav>

            {/* Main Content */}
            <main className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
                {/* Hero Section */}
                <div className="text-center mb-12">
                    <h1 className="text-4xl md:text-6xl font-bold text-gray-900 mb-4 leading-tight">
                        Shorten URLs with{" "}
                        <span className="bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
                            Style
                        </span>
                    </h1>
                    <p className="text-base sm:text-lg md:text-xl text-gray-600 max-w-2xl mx-auto">
                        Transform your long URLs into short, shareable links. Track clicks and analyze your audience
                        with beautiful analytics.
                    </p>
                </div>

                {/* URL Shortener Form */}
                <div className="bg-white rounded-2xl shadow-xl border border-gray-100 p-6 sm:p-8 mb-10">
                    <div className="space-y-4">
                        <div>
                            <label htmlFor="url" className="block text-sm font-medium text-gray-700 mb-2">
                                Enter your long URL
                            </label>
                            <div className="flex flex-col sm:flex-row gap-3 sm:gap-4">
                                <input
                                    id="url"
                                    type="url"
                                    value={url}
                                    onChange={(e) => setUrl(e.target.value)}
                                    placeholder="https://example.com/very/long/url/that/needs/shortening"
                                    className="flex-1 px-4 py-3 border border-gray-300 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all text-sm sm:text-base"
                                    required
                                />
                                <button
                                    onClick={handleCreateLink}
                                    disabled={isLoading}
                                    className="cursor-pointer px-6 sm:px-8 py-3 bg-gradient-to-r from-blue-600 to-purple-600 text-white rounded-xl hover:from-blue-700 hover:to-purple-700 transition-all duration-200 font-medium disabled:opacity-50 whitespace-nowrap text-sm sm:text-base"
                                >
                                    {isLoading ? "Shortening..." : "Shorten URL"}
                                </button>
                            </div>
                        </div>

                        {!isAuthenticated && (
                            <p className="text-amber-600 text-sm bg-amber-50 p-3 rounded-xl">
                                Please sign in to create short URLs
                            </p>
                        )}
                    </div>

                    {shortUrl && (
                        <div className="mt-6 p-4 bg-green-50 rounded-xl border border-green-200">
                            <p className="text-sm text-green-700 mb-2">Your shortened URL:</p>
                            <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-3">
                                <input
                                    type="text"
                                    value={shortUrl}
                                    readOnly
                                    className="flex-1 px-3 py-2 bg-white border border-green-300 rounded-lg text-green-800 font-mono text-sm sm:text-base"
                                />
                                <button
                                    onClick={() => copyToClipboard(`${baseUrl}/${shortUrl}`)}
                                    className="cursor-pointer px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors flex items-center justify-center gap-2 text-sm sm:text-base"
                                >
                                    <Copy size={16} />
                                    {copied ? "Copied!" : "Copy"}
                                </button>
                            </div>
                        </div>
                    )}
                </div>

                {/* Links List */}
                {isAuthenticated && (
                    <div className="bg-white rounded-2xl shadow-xl border border-gray-100 p-6 sm:p-8">
                        <h2 className="text-2xl font-bold text-gray-900 mb-6">Your Links</h2>

                        {links.length === 0 ? (
                            <div className="text-center py-12">
                                <Link2 size={48} className="text-gray-300 mx-auto mb-4" />
                                <p className="text-gray-500 text-lg">No links created yet</p>
                                <p className="text-gray-400">Start by shortening your first URL above</p>
                            </div>
                        ) : (
                            <div className="space-y-4">
                                {links.map((link) => (
                                    <div
                                        key={link.id}
                                        className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 p-4 border border-gray-200 rounded-xl hover:bg-gray-50 transition-colors"
                                    >
                                        <div className="flex-1 min-w-0">
                                            <div className="flex items-center gap-2 mb-2">
                                                <p className="font-mono text-blue-600 font-medium truncate">
                                                    /{link.short_token}
                                                </p>
                                                <button
                                                    onClick={() => copyToClipboard(`${baseUrl}/${link.short_token}`)}
                                                    className="cursor-pointer p-1 text-gray-400 hover:text-gray-600 transition-colors"
                                                >
                                                    <Copy size={16} />
                                                </button>
                                            </div>
                                            <p className="text-gray-600 truncate text-sm">{link.url}</p>
                                            <p className="text-gray-400 text-xs mt-1">
                                                Created {new Date(link.CreatedAt).toLocaleDateString()}
                                            </p>
                                        </div>

                                        <div className="flex items-center gap-2">
                                            <button
                                                onClick={() => window.open(`${baseUrl}/${link.short_token}`, "_blank")}
                                                className="cursor-pointer p-2 text-gray-400 hover:text-blue-600 transition-colors"
                                                title="Open link"
                                            >
                                                <ExternalLink size={18} />
                                            </button>
                                            <button
                                                onClick={() => handleShowStats(link.short_token)}
                                                className="cursor-pointer p-2 text-gray-400 hover:text-purple-600 transition-colors"
                                                title="View statistics"
                                            >
                                                <BarChart3 size={18} />
                                            </button>
                                            <button
                                                onClick={() => handleDeleteLink(link.short_token)}
                                                className="cursor-pointer p-2 text-gray-400 hover:text-purple-600 transition-colors"
                                                title="Delete link"
                                            >
                                                <Trash size={18} />
                                            </button>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        )}
                    </div>
                )}
            </main>

            {/* Modals */}
            <LoginModal
                isOpen={showLoginModal}
                onClose={() => setShowLoginModal(false)}
                onSwitchToRegister={() => {
                    setShowLoginModal(false);
                    setShowRegisterModal(true);
                }}
            />

            <RegisterModal
                isOpen={showRegisterModal}
                onClose={() => setShowRegisterModal(false)}
                onSwitchToLogin={() => {
                    setShowRegisterModal(false);
                    setShowLoginModal(true);
                }}
            />

            <StatsModal
                isOpen={showStatsModal}
                onClose={() => setShowStatsModal(false)}
                shortToken={selectedLinkToken}
            />
            <ToastContainer
                position="bottom-right"
                autoClose={3000}
                hideProgressBar={false}
                newestOnTop={false}
                closeOnClick
                rtl={false}
                pauseOnFocusLoss
                draggable
                pauseOnHover
            />
        </div>
    );
};

export default Index;
