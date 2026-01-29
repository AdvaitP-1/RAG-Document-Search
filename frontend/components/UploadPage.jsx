"use client";

import { useState } from "react";
import { apiRequest } from "../lib/apiClient";

export default function UploadPage({ accessToken }) {
  const [collectionId, setCollectionId] = useState("");
  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [status, setStatus] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (event) => {
    event.preventDefault();
    setLoading(true);
    setStatus("");
    try {
      const data = await apiRequest(`/collections/${collectionId}/docs`, {
        method: "POST",
        token: accessToken,
        body: { title, content },
      });
      setStatus(`Created doc ${data.document.id}, job ${data.job.id}`);
      setTitle("");
      setContent("");
    } catch (err) {
      setStatus(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="card">
      <h3>Upload a document</h3>
      <form onSubmit={handleSubmit}>
        <div className="field">
          <label>Collection ID</label>
          <input
            required
            value={collectionId}
            onChange={(event) => setCollectionId(event.target.value)}
            placeholder="collection uuid"
          />
        </div>
        <div className="field">
          <label>Title</label>
          <input
            required
            value={title}
            onChange={(event) => setTitle(event.target.value)}
            placeholder="Notes on onboarding"
          />
        </div>
        <div className="field">
          <label>Content (plain text/markdown)</label>
          <textarea
            required
            value={content}
            onChange={(event) => setContent(event.target.value)}
            placeholder="Paste your document here."
          />
        </div>
        <button className="btn" type="submit" disabled={loading}>
          {loading ? "Uploading..." : "Upload"}
        </button>
      </form>
      {status && <p className="muted">{status}</p>}
    </div>
  );
}
