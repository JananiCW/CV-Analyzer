import { useState, useEffect } from "react";
import "./App.css";

const API = "http://localhost:8081";

function App() {
  const [file, setFile] = useState(null);
  const [analysis, setAnalysis] = useState("");
  const [history, setHistory] = useState([]);
  const [loading, setLoading] = useState(false);
  const [filename, setFilename] = useState("");

  useEffect(() => {
    fetchHistory();
  }, []);

  // 📥 Fetch history from backend
  const fetchHistory = async () => {
    try {
      const res = await fetch(`${API}/history`);
      const data = await res.json();
      setHistory(data || []);
    } catch (err) {
      console.error("Error fetching history:", err);
    }
  };

  // 📄 Handle file selection
  const handleFileChange = (e) => {
    setFile(e.target.files[0]);
    setAnalysis("");
  };

  // 🚀 Upload and analyze CV
  const analyzeCV = async () => {
    if (!file) return alert("Please select a PDF file first!");

    setLoading(true);
    setAnalysis("");

    const formData = new FormData();
    formData.append("cv", file);

    try {
      const res = await fetch(`${API}/analyze`, {
        method: "POST",
        body: formData,
      });
      const data = await res.json();

      if (data.error) {
        alert("Error: " + data.error);
      } else {
        setAnalysis(data.analysis);
        setFilename(data.filename);
        fetchHistory();
      }
    } catch (err) {
      alert("Failed to connect to server!");
      console.error(err);
    }

    setLoading(false);
  };

  // 🗑️ Delete a history item
  const deleteHistory = async (id) => {
    try {
      await fetch(`${API}/history/${id}`, { method: "DELETE" });
      fetchHistory();
    } catch (err) {
      console.error("Error deleting:", err);
    }
  };

  // 📊 Format analysis text with line breaks
  const formatAnalysis = (text) => {
    return text.split("\n").map((line, i) => (
      <p key={i} className={line.startsWith("#") ? "section-title" : "section-text"}>
        {line}
      </p>
    ));
  };

  return (
    <div className="app">
      <div className="container">

        {/* ── Header ── */}
        <div className="header">
          <h1>🎯 AI CV Analyzer</h1>
          <p className="subtitle">Upload your CV and get instant AI feedback</p>
        </div>

        {/* ── Upload Section ── */}
        <div className="upload-section">
          <div className="upload-box">
            <p>📄 Select your CV (PDF only)</p>
            <input
              type="file"
              accept=".pdf"
              onChange={handleFileChange}
              className="file-input"
            />
            {file && <p className="file-name">✅ {file.name}</p>}
          </div>

          <button
            className="analyze-btn"
            onClick={analyzeCV}
            disabled={loading || !file}
          >
            {loading ? "🤖 Analyzing..." : "🚀 Analyze CV"}
          </button>
        </div>

        {/* ── Analysis Result ── */}
        {analysis && (
          <div className="result-section">
            <h2>📊 Analysis Result for: {filename}</h2>
            <div className="analysis-box">
              {formatAnalysis(analysis)}
            </div>
          </div>
        )}

        {/* ── History Section ── */}
        {history && history.length > 0 && (
          <div className="history-section">
            <h2>📁 Previous Analyses</h2>
            <ul className="history-list">
              {history.map((item) => (
                <li key={item.id} className="history-item">
                  <div className="history-info">
                    <span className="history-filename">📄 {item.filename}</span>
                    <span className="history-date">
                      {new Date(item.createdAt).toLocaleDateString()}
                    </span>
                  </div>
                  <div className="history-actions">
                    <button
                      className="view-btn"
                      onClick={() => {
                        setAnalysis(item.analysis);
                        setFilename(item.filename);
                      }}
                    >
                      👁️ View
                    </button>
                    <button
                      className="del-btn"
                      onClick={() => deleteHistory(item.id)}
                    >
                      🗑️
                    </button>
                  </div>
                </li>
              ))}
            </ul>
          </div>
        )}

        <p className="footer">Powered by Google Gemini AI + Go + MongoDB 🚀</p>
      </div>
    </div>
  );
}

export default App;