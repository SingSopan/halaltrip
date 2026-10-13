"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState, type FormEvent } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { login, register } from "@/src/services";
import { useAuthStore } from "@/src/stores/use-auth-store";
import { APIError } from "@/src/services/client";

export function AuthPage({ initialMode = "login" }: { initialMode?: "login" | "register" }) {
  const [mode, setMode] = useState(initialMode);
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [validationMessage, setValidationMessage] = useState("");
  const router = useRouter();
  const queryClient = useQueryClient();
  const setSession = useAuthStore((state) => state.setSession);
  const mutation = useMutation({
    mutationFn: async () => mode === "register" ? register({ name: name.trim(), email: email.trim(), password }) : login({ email: email.trim(), password }),
    onSuccess: (response) => {
      setSession(response.access_token, response.user);
      queryClient.clear();
      router.push("/trips");
    },
  });

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setValidationMessage("");
    if (mode === "register" && !name.trim()) { setValidationMessage("Nama wajib diisi."); return; }
    if (!email.includes("@")) { setValidationMessage("Masukkan email yang valid."); return; }
    const passwordBytes = new TextEncoder().encode(password).length;
    if (passwordBytes < (mode === "register" ? 8 : 1)) {
      setValidationMessage(mode === "register" ? "Password minimal 8 karakter." : "Password wajib diisi.");
      return;
    }
    if (mode === "register" && passwordBytes > 72) {
      setValidationMessage("Password maksimal 72 byte. Gunakan password yang lebih pendek.");
      return;
    }
    mutation.mutate();
  }

  const serverError = mutation.error instanceof APIError
    ? mutation.error.status >= 500 && mode === "register"
      ? `${mutation.error.message} Periksa log backend, koneksi PostgreSQL, dan pastikan migrasi database sudah dijalankan.`
      : mutation.error.message
    : mutation.error
      ? "Backend tidak dapat dijangkau. Pastikan API berjalan di http://localhost:8080 dan CORS mengizinkan origin frontend."
      : "";

  return (
    <section className="auth-panel">
      <div className="eyebrow">YOUR HALALTRIP ACCOUNT</div>
      <h1>{mode === "login" ? "Welcome back." : "Create your account."}</h1>
      <p>Masuk untuk menyimpan preferensi, tempat, dan rencana perjalanan di backend HalalTrip.</p>
      <form className="auth-form" onSubmit={submit}>
        {mode === "register" && <label className="form-label">Nama<input autoComplete="name" className="form-input" value={name} onChange={(event) => setName(event.target.value)} maxLength={100} required /></label>}
        <label className="form-label">Email<input type="email" autoComplete="email" className="form-input" value={email} onChange={(event) => setEmail(event.target.value)} maxLength={254} required /></label>
        <label className="form-label">Password<input type="password" autoComplete={mode === "login" ? "current-password" : "new-password"} className="form-input" value={password} onChange={(event) => setPassword(event.target.value)} minLength={mode === "register" ? 8 : 1} maxLength={72} required /></label>
        {(validationMessage || mutation.error) && <p className="error-text" role="alert">{validationMessage || serverError}</p>}
        <button className="button primary" type="submit" disabled={mutation.isPending}>{mutation.isPending ? "Mengirim..." : mode === "login" ? "Masuk" : "Daftar"}</button>
      </form>
      <p className="auth-switch">{mode === "login" ? "Belum punya akun?" : "Sudah punya akun?"} <button type="button" onClick={() => { setMode(mode === "login" ? "register" : "login"); mutation.reset(); setValidationMessage(""); }}>{mode === "login" ? "Daftar" : "Masuk"}</button></p>
      <small>Token sesi hanya disimpan di memori browser dan hilang saat halaman dimuat ulang.</small>
      <Link href="/" className="back-link">Kembali ke beranda</Link>
    </section>
  );
}
