"use client";

import Link from "next/link";
import { supabase } from "../lib/supabaseClient";

export default function AppHeader({ email }) {
  const handleSignOut = async () => {
    await supabase.auth.signOut();
  };

  return (
    <div className="header">
      <div>
        <h2>RAG Document Search</h2>
        <p className="muted">{email}</p>
      </div>
      <div className="nav">
        <Link href="/collections">Collections</Link>
        <Link href="/upload">Upload</Link>
        <button className="btn secondary" type="button" onClick={handleSignOut}>
          Sign out
        </button>
      </div>
    </div>
  );
}
