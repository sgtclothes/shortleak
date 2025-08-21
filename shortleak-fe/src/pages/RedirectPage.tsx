// src/pages/RedirectPage.jsx
import { useEffect } from "react";
import { useParams } from "react-router-dom";
import Redirecting from "../components/Redirecting";

const baseUrlAPI = import.meta.env.VITE_BASE_URL_API ?? "http://173.249.46.3:8090";

export default function RedirectPage() {
    const { shortToken } = useParams();

    useEffect(() => {
        const fetchUrl = async () => {
            try {
                const res = await fetch(`${baseUrlAPI}/api/links/${shortToken}`);
                const data = await res.json();
                if (data.url) {
                    window.location.href = baseUrlAPI + "/" + shortToken;
                } else {
                    window.location.href = "/404";
                }
            } catch (err) {
                console.error(err);
                window.location.href = "/404";
            }
        };

        fetchUrl();
    }, [shortToken]);

    return (
      <Redirecting shortToken={shortToken ?? ""} originalUrl={baseUrlAPI + "/" + shortToken} onRetry={() => window.location.reload()} />
    );
}
