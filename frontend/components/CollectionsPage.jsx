"use client";

import { useEffect, useState } from "react";
import { apiRequest } from "../lib/apiClient";

export default function CollectionsPage({ accessToken }) {
  const [collections, setCollections] = useState([]);
  const [name, setName] = useState("");
  const [status, setStatus] = useState("");
  const [loading, setLoading] = useState(false);

  const loadCollections = async () => {
    setLoading(true);
    setStatus("");
    try {
      const data = await apiRequest("/collections", { token: accessToken });
      setCollections(data);
    } catch (err) {
      setStatus(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (accessToken) {
      loadCollections();
    }
  }, [accessToken]);

  const createCollection = async (event) => {
    event.preventDefault();
    if (!name.trim()) return;
    setLoading(true);
    setStatus("");
    try {
      await apiRequest("/collections", {
        method: "POST",
        token: accessToken,
        body: { name },
      });
      setName("");
      await loadCollections();
    } catch (err) {
      setStatus(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="card">
      <h3>Your collections</h3>
      <form onSubmit={createCollection} className="row">
        <div className="field" style={{ flex: 1 }}>
          <label>Collection name</label>
          <input
            value={name}
            onChange={(event) => setName(event.target.value)}
            placeholder="Product docs"
          />
        </div>
        <button className="btn" type="submit" disabled={loading}>
          Create
        </button>
      </form>
      {status && <p className="muted">{status}</p>}
      <div className="row">
        {collections.map((collection) => (
          <div key={collection.id} className="card" style={{ flex: 1 }}>
            <h4>{collection.name}</h4>
            <p className="muted">{collection.id}</p>
          </div>
        ))}
      </div>
      {!loading && collections.length === 0 && (
        <p className="muted">No collections yet.</p>
      )}
    </div>
  );
}
