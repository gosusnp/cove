// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { Capacitor } from "@capacitor/core";
import { SocialLogin } from "@capgo/capacitor-social-login";
import { useEffect, useState } from "preact/hooks";
import { useLocation } from "preact-iso";
import { Button } from "../components/ui/Button.jsx";
import { PageTitle } from "../components/ui/PageTitle.jsx";
import { useAuth } from "../Auth.jsx";
import { apiFetch } from "../lib/api.js";

function GoogleIcon() {
	return (
		<svg
			xmlns="http://www.w3.org/2000/svg"
			viewBox="0 0 48 48"
			width="18"
			height="18"
			role="img"
			aria-label="Google"
		>
			<path
				fill="#EA4335"
				d="M24 9.5c3.54 0 6.71 1.22 9.21 3.6l6.85-6.85C35.9 2.38 30.47 0 24 0 14.62 0 6.51 5.38 2.56 13.22l7.98 6.19C12.43 13.72 17.74 9.5 24 9.5z"
			/>
			<path
				fill="#4285F4"
				d="M46.98 24.55c0-1.57-.15-3.09-.38-4.55H24v9.02h12.94c-.58 2.96-2.26 5.48-4.78 7.18l7.73 6c4.51-4.18 7.09-10.36 7.09-17.65z"
			/>
			<path
				fill="#FBBC05"
				d="M10.53 28.59c-.48-1.45-.76-2.99-.76-4.59s.27-3.14.76-4.59l-7.98-6.19C.92 16.46 0 20.12 0 24c0 3.88.92 7.54 2.56 10.78l7.97-6.19z"
			/>
			<path
				fill="#34A853"
				d="M24 48c6.48 0 11.93-2.13 15.89-5.81l-7.73-6c-2.15 1.45-4.92 2.3-8.16 2.3-6.26 0-11.57-4.22-13.47-9.91l-7.98 6.19C6.51 42.62 14.62 48 24 48z"
			/>
		</svg>
	);
}

export function Login() {
	const { route } = useLocation();
	const { updateUser } = useAuth();
	const isNative = Capacitor.isNativePlatform();
	const [ready, setReady] = useState(!isNative);
	const [error, setError] = useState(null);

	useEffect(() => {
		if (!isNative) return;
		SocialLogin.initialize({
			google: { webClientId: import.meta.env.VITE_GOOGLE_CLIENT_ID },
		})
			.then(() => setReady(true))
			.catch((e) => console.error("SocialLogin.initialize failed:", e));
	}, [isNative]);

	async function signInWithGoogle() {
		setError(null);
		try {
			const result = await SocialLogin.login({
				provider: "google",
				options: {},
			});
			const tokenRes = await apiFetch("/auth/google/token", {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({ id_token: result.result.idToken }),
			});
			if (!tokenRes.ok) throw new Error(`auth failed: ${tokenRes.status}`);
			const me = await apiFetch("/api/users/me");
			if (me.ok) {
				updateUser(await me.json());
				route("/");
			}
		} catch (e) {
			console.error("signInWithGoogle failed:", e);
			setError("Sign-in failed. Please try again.");
		}
	}

	return (
		<main class="flex min-h-dvh flex-col items-center justify-center gap-8 px-4">
			<div class="flex flex-col items-center gap-2">
				<PageTitle class="text-6xl tracking-tight">Cove</PageTitle>
				<p class="text-sm" style={{ color: "var(--color-muted)" }}>
					Your space.
				</p>
			</div>

			{error && (
				<p class="text-sm" style={{ color: "var(--color-error)" }}>
					{error}
				</p>
			)}

			{isNative ? (
				<Button
					variant="outline"
					size="lg"
					onClick={signInWithGoogle}
					disabled={!ready}
				>
					<GoogleIcon />
					Continue with Google
				</Button>
			) : (
				<Button
					variant="outline"
					size="lg"
					onClick={() => {
						window.location.href = "/auth/login";
					}}
				>
					<GoogleIcon />
					Continue with Google
				</Button>
			)}
		</main>
	);
}
