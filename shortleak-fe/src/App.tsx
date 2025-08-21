import { BrowserRouter, Routes, Route } from "react-router-dom";
import RedirectPage from "./pages/RedirectPage";
import Index from "./pages/Index";
import NotFound from "./components/NotFound";

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Index />} />
        <Route path="/:shortToken" element={<RedirectPage />} />
        <Route path="/404" element={<NotFound />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
