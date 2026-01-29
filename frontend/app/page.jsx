"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import AuthPage from "../components/AuthPage";
import { useSession } from "../lib/useSession";

export default function HomePage() {
  const { session, loading } = useSession();
  const router = useRouter();

  useEffect(() => {
    if (session) {
      router.push("/collections");
    }
  }, [session, router]);

  if (loading) {
    return <div className="card">Loading...</div>;
  }

  if (!session) {
    return <AuthPage />;
  }

  return <div className="card">Redirecting to collections...</div>;
}
