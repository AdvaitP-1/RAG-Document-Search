"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import AppHeader from "../../components/AppHeader";
import UploadPage from "../../components/UploadPage";
import { useSession } from "../../lib/useSession";

export default function UploadRoute() {
  const { session, loading } = useSession();
  const router = useRouter();

  useEffect(() => {
    if (!loading && !session) {
      router.push("/");
    }
  }, [loading, session, router]);

  if (loading) {
    return <div className="card">Loading...</div>;
  }

  if (!session) {
    return <div className="card">Redirecting to sign in...</div>;
  }

  return (
    <>
      <AppHeader email={session.user.email} />
      <UploadPage accessToken={session.access_token} />
    </>
  );
}
