<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import * as Card from '$lib/components/ui/card';
	import { notificationStore } from '$lib/services/notificationService';
	import { messageStorage } from '$lib/services/messageStorageService';

	interface TestResult {
		message: string;
		timestamp: string;
	}

	interface DebugInfo {
		database: {
			total: number;
			unread: number;
			byAudience: Record<string, number>;
			oldestMessage?: string;
			newestMessage?: string;
		};
		memory: {
			notifications: number;
			unreadCount: number;
			lastFetched: string | null;
			isLoading: boolean;
		};
		performance: {
			memoryVsDatabase: {
				memory: number;
				database: number;
				synced: boolean;
			};
		};
	}

	let debugInfo = $state<DebugInfo | null>(null);
	let isLoading = $state(false);
	let testResults = $state<TestResult[]>([]);

	onMount(() => {
		loadDebugInfo();
	});

	async function loadDebugInfo() {
		isLoading = true;
		try {
			debugInfo = await notificationStore.getDebugInfo();
		} catch (error) {
			console.error('Failed to load debug info:', error);
		} finally {
			isLoading = false;
		}
	}

	async function testDexieOperations() {
		isLoading = true;
		testResults = [];

		try {
			// Test 1: Basic storage
			addTestResult('Testing basic message storage...');
			const testMessage = {
				id: Date.now(),
				type: 'broadcast' as const,
				title: 'Test Message',
				message: 'This is a test message from Dexie debug page',
				timestamp: new Date().toISOString(),
				read: false,
				data: { audience: 'test' }
			};

			await messageStorage.storeMessages([testMessage]);
			addTestResult('✅ Message stored successfully');

			// Test 2: Retrieval
			addTestResult('Testing message retrieval...');
			const messages = await messageStorage.getMessages();
			addTestResult(`✅ Retrieved ${messages.length} messages`);

			// Test 3: Mark as read
			addTestResult('Testing mark as read...');
			await messageStorage.markAsRead(testMessage.id);
			addTestResult('✅ Message marked as read');

			// Test 4: Search
			addTestResult('Testing search functionality...');
			const searchResults = await messageStorage.searchMessages('test');
			addTestResult(`✅ Search found ${searchResults.length} results`);

			// Test 5: Stats
			addTestResult('Testing stats...');
			const stats = await messageStorage.getStats();
			addTestResult(`✅ Stats: ${stats.total} total, ${stats.unread} unread`);

			// Refresh debug info
			await loadDebugInfo();
		} catch (error) {
			const errorMessage = error instanceof Error ? error.message : 'Unknown error';
			addTestResult(`❌ Error: ${errorMessage}`);
		} finally {
			isLoading = false;
		}
	}

	function addTestResult(message: string) {
		testResults = [
			...testResults,
			{
				message,
				timestamp: new Date().toLocaleTimeString()
			}
		];
	}

	async function clearDatabase() {
		if (confirm('Are you sure you want to clear all messages from IndexedDB?')) {
			try {
				isLoading = true;
				// Clear all messages
				const messages = await messageStorage.getMessages();
				for (const msg of messages) {
					await messageStorage.markAsRead(msg.id); // This will help test the delete
				}

				// Clear old messages (all of them)
				await messageStorage.clearOldMessages(0);

				addTestResult('✅ Database cleared');
				await loadDebugInfo();
			} catch (error) {
				const errorMessage = error instanceof Error ? error.message : 'Unknown error';
				addTestResult(`❌ Clear failed: ${errorMessage}`);
			} finally {
				isLoading = false;
			}
		}
	}
</script>

<svelte:head>
	<title>Dexie Debug - Night Owls</title>
</svelte:head>

<div class="container mx-auto p-6 space-y-6">
	<div class="flex items-center justify-between">
		<h1 class="text-2xl font-bold">Dexie.js Debug Page</h1>
		<div class="flex gap-2">
			<Button onclick={loadDebugInfo} disabled={isLoading}>Refresh Info</Button>
			<Button onclick={testDexieOperations} disabled={isLoading}>Run Tests</Button>
			<Button variant="destructive" onclick={clearDatabase} disabled={isLoading}>Clear DB</Button>
		</div>
	</div>

	{#if isLoading}
		<Card.Root>
			<Card.Content class="p-6 text-center">
				<div
					class="animate-spin h-8 w-8 border-4 border-primary border-t-transparent rounded-full mx-auto mb-4"
				></div>
				<p>Loading...</p>
			</Card.Content>
		</Card.Root>
	{/if}

	{#if debugInfo}
		<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
			<!-- Database Stats -->
			<Card.Root>
				<Card.Header>
					<Card.Title>Database Statistics</Card.Title>
				</Card.Header>
				<Card.Content class="space-y-2">
					<div class="flex justify-between">
						<span>Total Messages:</span>
						<span class="font-mono">{debugInfo.database.total}</span>
					</div>
					<div class="flex justify-between">
						<span>Unread Messages:</span>
						<span class="font-mono">{debugInfo.database.unread}</span>
					</div>
					<div class="flex justify-between">
						<span>Oldest Message:</span>
						<span class="font-mono text-xs">
							{debugInfo.database.oldestMessage
								? new Date(debugInfo.database.oldestMessage).toLocaleString()
								: 'None'}
						</span>
					</div>
					<div class="flex justify-between">
						<span>Newest Message:</span>
						<span class="font-mono text-xs">
							{debugInfo.database.newestMessage
								? new Date(debugInfo.database.newestMessage).toLocaleString()
								: 'None'}
						</span>
					</div>
				</Card.Content>
			</Card.Root>

			<!-- Memory State -->
			<Card.Root>
				<Card.Header>
					<Card.Title>Memory State</Card.Title>
				</Card.Header>
				<Card.Content class="space-y-2">
					<div class="flex justify-between">
						<span>In-Memory Messages:</span>
						<span class="font-mono">{debugInfo.memory.notifications}</span>
					</div>
					<div class="flex justify-between">
						<span>Unread Count:</span>
						<span class="font-mono">{debugInfo.memory.unreadCount}</span>
					</div>
					<div class="flex justify-between">
						<span>Last Fetched:</span>
						<span class="font-mono text-xs">
							{debugInfo.memory.lastFetched
								? new Date(debugInfo.memory.lastFetched).toLocaleString()
								: 'Never'}
						</span>
					</div>
					<div class="flex justify-between">
						<span>Is Loading:</span>
						<span class="font-mono">{debugInfo.memory.isLoading ? 'Yes' : 'No'}</span>
					</div>
				</Card.Content>
			</Card.Root>
		</div>

		<!-- Audience Breakdown -->
		<Card.Root>
			<Card.Header>
				<Card.Title>Messages by Audience</Card.Title>
			</Card.Header>
			<Card.Content>
				{#if Object.keys(debugInfo.database.byAudience).length > 0}
					<div class="grid grid-cols-2 md:grid-cols-4 gap-4">
						{#each Object.entries(debugInfo.database.byAudience) as [audience, count] (audience)}
							<div class="text-center p-3 bg-muted rounded-lg">
								<div class="font-bold text-lg">{count}</div>
								<div class="text-sm text-muted-foreground capitalize">{audience}</div>
							</div>
						{/each}
					</div>
				{:else}
					<p class="text-muted-foreground text-center py-4">No messages found</p>
				{/if}
			</Card.Content>
		</Card.Root>

		<!-- Sync Status -->
		<Card.Root>
			<Card.Header>
				<Card.Title>Synchronization Status</Card.Title>
			</Card.Header>
			<Card.Content>
				<div class="flex items-center justify-between p-4 bg-muted rounded-lg">
					<span>Memory ↔ Database Sync:</span>
					<span
						class="font-mono {debugInfo.performance.memoryVsDatabase.synced
							? 'text-green-600'
							: 'text-red-600'}"
					>
						{debugInfo.performance.memoryVsDatabase.synced ? '✅ Synced' : '❌ Out of Sync'}
					</span>
				</div>
				<div class="mt-2 text-sm text-muted-foreground">
					Memory: {debugInfo.performance.memoryVsDatabase.memory} | Database: {debugInfo.performance
						.memoryVsDatabase.database}
				</div>
			</Card.Content>
		</Card.Root>
	{/if}

	<!-- Test Results -->
	{#if testResults.length > 0}
		<Card.Root>
			<Card.Header>
				<Card.Title>Test Results</Card.Title>
			</Card.Header>
			<Card.Content>
				<div class="space-y-2 max-h-64 overflow-y-auto">
					{#each testResults as result, i (i)}
						<div class="flex justify-between items-center p-2 bg-muted rounded text-sm">
							<span>{result.message}</span>
							<span class="text-xs text-muted-foreground">{result.timestamp}</span>
						</div>
					{/each}
				</div>
			</Card.Content>
		</Card.Root>
	{/if}
</div>
