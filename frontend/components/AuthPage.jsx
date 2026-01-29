"use client";

import { useState } from "react";
import { supabase } from "../lib/supabaseClient";

export default function AuthPage() {
  const [mode, setMode] = useState("login");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [status, setStatus] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (event) => {
    event.preventDefault();
    setLoading(true);
    setStatus("");
    try {
      if (mode === "signup") {
        const { error } = await supabase.auth.signUp({ email, password });
        if (error) throw error;
        setStatus("Check your email to confirm signup.");
      } else {
        const { error } = await supabase.auth.signInWithPassword({
          email,
          password,
        });
        if (error) throw error;
        setStatus("Signed in.");
      }
    } catch (err) {
      setStatus(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="card">
      <h2>{mode === "signup" ? "Create account" : "Sign in"}</h2>
      <form onSubmit={handleSubmit}>
        <div className="field">
          <label>Email</label>
          <input
            type="email"
            required
            value={email}
            onChange={(event) => setEmail(event.target.value)}
            placeholder="you@example.com"
          />
        </div>
        <div className="field">
          <label>Password</label>
          <input
            type="password"
            required
            value={password}
            onChange={(event) => setPassword(event.target.value)}
            placeholder="••••••••"
          />
        </div>
        <button className="btn" type="submit" disabled={loading}>
          {loading ? "Working..." : mode === "signup" ? "Sign up" : "Sign in"}
        </button>
      </form>
      {status && <p className="muted">{status}</p>}
      <button
        className="btn secondary"
        type="button"
        onClick={() => setMode(mode === "signup" ? "login" : "signup")}
      >
        {mode === "signup" ? "Have an account? Sign in" : "New here? Sign up"}
      </button>
    </div>
  );
}
