// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useSignal } from "@preact/signals";
import { useEffect } from "preact/hooks";
import { Button } from "../components/ui/Button.jsx";
import { ConfirmDialog } from "../components/ui/ConfirmDialog.jsx";
import {
	Dialog,
	DialogClose,
	DialogContent,
	DialogDescription,
	DialogTitle,
} from "../components/ui/Dialog.jsx";
import { ListDetail } from "../components/ui/ListDetail.jsx";
import { ListItem } from "../components/ui/ListItem.jsx";
import { Row, Section } from "../components/ui/Section.jsx";
import { TextField } from "../components/ui/TextField.jsx";
import { useDialog } from "../hooks/useDialog.js";
import { apiFetch } from "../lib/api.js";

export function AdminServiceAccounts() {
	const accounts = useSignal([]);
	const accountsLoading = useSignal(true);
	const selectedId = useSignal(null);
	const tokens = useSignal([]);
	const tokensLoading = useSignal(false);

	// Create service account
	const createSADialog = useDialog();
	const saName = useSignal("");
	const creatingSA = useSignal(false);
	const createSAError = useSignal("");

	// Delete service account
	const deleteSADialog = useDialog();
	const pendingDeleteId = useSignal(null);

	// Create token
	const createTokenDialog = useDialog();
	const tokenName = useSignal("");
	const creatingToken = useSignal(false);
	const createTokenError = useSignal("");
	const createdToken = useSignal(null);
	const copied = useSignal(false);

	useEffect(() => {
		apiFetch("/api/admin/service-accounts")
			.then((r) => (r.ok ? r.json() : []))
			.then((data) => {
				accounts.value = data ?? [];
			})
			.catch(() => {})
			.finally(() => {
				accountsLoading.value = false;
			});
	}, []);

	useEffect(() => {
		if (!selectedId.value) {
			tokens.value = [];
			return;
		}
		tokensLoading.value = true;
		apiFetch(`/api/admin/service-accounts/${selectedId.value}/tokens`)
			.then((r) => (r.ok ? r.json() : []))
			.then((data) => {
				tokens.value = data ?? [];
			})
			.catch(() => {})
			.finally(() => {
				tokensLoading.value = false;
			});
	}, [selectedId.value]);

	function openCreateSA() {
		saName.value = "";
		createSAError.value = "";
		createSADialog.show();
	}

	async function handleCreateSA(e) {
		e.preventDefault();
		const name = saName.value.trim();
		if (!name) {
			createSAError.value = "Name is required.";
			return;
		}
		creatingSA.value = true;
		createSAError.value = "";
		try {
			const r = await apiFetch("/api/admin/service-accounts", {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({ name }),
			});
			if (!r.ok) {
				createSAError.value = "Failed to create service account.";
				return;
			}
			const data = await r.json();
			accounts.value = [...accounts.value, data];
			selectedId.value = data.id;
			createSADialog.hide();
		} catch {
			createSAError.value = "Failed to create service account.";
		} finally {
			creatingSA.value = false;
		}
	}

	function confirmDeleteSA(id) {
		pendingDeleteId.value = id;
		deleteSADialog.show();
	}

	async function handleDeleteSA() {
		const id = pendingDeleteId.value;
		if (!id) return;
		const r = await apiFetch(`/api/admin/service-accounts/${id}`, {
			method: "DELETE",
		});
		if (r.ok) {
			accounts.value = accounts.value.filter((a) => a.id !== id);
			if (selectedId.value === id) {
				selectedId.value = null;
			}
		}
		pendingDeleteId.value = null;
	}

	function openCreateToken() {
		tokenName.value = "";
		createTokenError.value = "";
		createdToken.value = null;
		createTokenDialog.show();
	}

	async function handleCreateToken(e) {
		e.preventDefault();
		const name = tokenName.value.trim();
		if (!name) {
			createTokenError.value = "Token name is required.";
			return;
		}
		creatingToken.value = true;
		createTokenError.value = "";
		try {
			const r = await apiFetch(
				`/api/admin/service-accounts/${selectedId.value}/tokens`,
				{
					method: "POST",
					headers: { "Content-Type": "application/json" },
					body: JSON.stringify({ name }),
				},
			);
			if (!r.ok) {
				createTokenError.value = "Failed to create token.";
				return;
			}
			const data = await r.json();
			tokens.value = [
				...tokens.value,
				{ id: data.id, name: data.name, created_at: data.created_at },
			];
			createdToken.value = data.token;
		} catch {
			createTokenError.value = "Failed to create token.";
		} finally {
			creatingToken.value = false;
		}
	}

	function handleCopy() {
		navigator.clipboard.writeText(createdToken.value);
		copied.value = true;
		setTimeout(() => {
			copied.value = false;
		}, 2000);
	}

	async function handleDeleteToken(tokenId) {
		const r = await apiFetch(
			`/api/admin/service-accounts/${selectedId.value}/tokens/${tokenId}`,
			{ method: "DELETE" },
		);
		if (r.ok) {
			tokens.value = tokens.value.filter((t) => t.id !== tokenId);
		}
	}

	const selectedAccount = accounts.value.find((a) => a.id === selectedId.value);

	const list = (
		<div class="flex flex-col">
			<div
				class="flex items-center justify-between px-4 py-3 border-b"
				style={{ borderColor: "var(--color-border)" }}
			>
				<h2
					class="text-xs font-semibold uppercase tracking-widest"
					style={{ color: "var(--color-muted)" }}
				>
					Service Accounts
				</h2>
				<Button
					variant="primary"
					size="sm"
					onClick={openCreateSA}
					disabled={accountsLoading.value}
				>
					+ New
				</Button>
			</div>
			{!accountsLoading.value && accounts.value.length === 0 && (
				<p class="px-4 py-6 text-sm" style={{ color: "var(--color-muted)" }}>
					No service accounts yet.
				</p>
			)}
			{accounts.value.map((a, i) => (
				<ListItem
					key={a.id}
					label={a.name}
					sublabel={`Created ${new Date(a.created_at).toLocaleDateString()}`}
					active={selectedId.value === a.id}
					isLast={i === accounts.value.length - 1}
					onClick={() => {
						selectedId.value = a.id;
					}}
					actions={
						<Button
							variant="destructive"
							size="sm"
							onClick={(e) => {
								e.stopPropagation();
								confirmDeleteSA(a.id);
							}}
						>
							Delete
						</Button>
					}
				/>
			))}
		</div>
	);

	const detail = selectedAccount ? (
		<div class="p-6">
			<Section
				title={selectedAccount.name}
				action={
					<Button
						variant="outline"
						size="sm"
						onClick={openCreateToken}
						disabled={tokensLoading.value}
					>
						Generate token
					</Button>
				}
			>
				{tokens.value.length === 0 && !tokensLoading.value ? (
					<Row label="No tokens yet" last />
				) : (
					tokens.value.map((t, i) => (
						<Row
							key={t.id}
							label={t.name}
							sublabel={`Created ${new Date(t.created_at).toLocaleDateString()}`}
							last={i === tokens.value.length - 1}
						>
							<Button
								variant="destructive"
								size="sm"
								onClick={() => handleDeleteToken(t.id)}
							>
								Revoke
							</Button>
						</Row>
					))
				)}
			</Section>
		</div>
	) : null;

	return (
		<>
			<ListDetail
				list={list}
				detail={detail}
				emptyState="Select a service account to manage its tokens."
				hasDetail={!!selectedAccount}
			/>

			{/* Create service account */}
			<Dialog openSignal={createSADialog.open}>
				<DialogContent>
					<form onSubmit={handleCreateSA}>
						<DialogTitle>New service account</DialogTitle>
						<DialogDescription>
							Enter a name to identify this service account.
						</DialogDescription>
						<div class="mt-4">
							<TextField
								id="sa-name"
								label="Name"
								placeholder="e.g. CI Bot"
								value={saName.value}
								onInput={(e) => {
									saName.value = e.target.value;
								}}
								autoFocus
							/>
							{createSAError.value && (
								<p class="text-sm mt-2" style={{ color: "var(--color-error)" }}>
									{createSAError.value}
								</p>
							)}
						</div>
						<div class="mt-6 flex justify-end gap-2">
							<DialogClose>
								<Button variant="ghost" size="sm" type="button">
									Cancel
								</Button>
							</DialogClose>
							<Button size="sm" type="submit" disabled={creatingSA.value}>
								{creatingSA.value ? "Creating…" : "Create"}
							</Button>
						</div>
					</form>
				</DialogContent>
			</Dialog>

			{/* Delete service account */}
			<ConfirmDialog
				openSignal={deleteSADialog.open}
				title="Delete service account?"
				description="This will permanently delete the service account and all its tokens. Any integrations using these tokens will stop working."
				confirmLabel="Delete"
				onConfirm={handleDeleteSA}
			/>

			{/* Create token */}
			<Dialog openSignal={createTokenDialog.open}>
				<DialogContent>
					{createdToken.value === null ? (
						<form onSubmit={handleCreateToken}>
							<DialogTitle>New API token</DialogTitle>
							<DialogDescription>
								Enter a name to identify this token.
							</DialogDescription>
							<div class="mt-4">
								<TextField
									id="token-name"
									label="Token name"
									placeholder="e.g. CI pipeline"
									value={tokenName.value}
									onInput={(e) => {
										tokenName.value = e.target.value;
									}}
									autoFocus
								/>
								{createTokenError.value && (
									<p
										class="text-sm mt-2"
										style={{ color: "var(--color-error)" }}
									>
										{createTokenError.value}
									</p>
								)}
							</div>
							<div class="mt-6 flex justify-end gap-2">
								<DialogClose>
									<Button variant="ghost" size="sm" type="button">
										Cancel
									</Button>
								</DialogClose>
								<Button size="sm" type="submit" disabled={creatingToken.value}>
									{creatingToken.value ? "Creating…" : "Create token"}
								</Button>
							</div>
						</form>
					) : (
						<>
							<DialogTitle>Token created</DialogTitle>
							<DialogDescription>
								Copy this token now — it won't be shown again.
							</DialogDescription>
							<div class="mt-4 flex flex-col gap-1.5">
								<label
									for="created-token"
									class="text-sm font-medium"
									style={{ color: "var(--color-text)" }}
								>
									Your new token
								</label>
								<div class="flex gap-2">
									<TextField
										id="created-token"
										value={createdToken.value}
										readOnly
										class="font-mono text-xs"
									/>
									<Button
										variant="outline"
										onClick={handleCopy}
										class="shrink-0"
									>
										{copied.value ? "Copied!" : "Copy"}
									</Button>
								</div>
							</div>
							<div class="mt-6 flex justify-end">
								<DialogClose>
									<Button size="sm">Done</Button>
								</DialogClose>
							</div>
						</>
					)}
				</DialogContent>
			</Dialog>
		</>
	);
}
