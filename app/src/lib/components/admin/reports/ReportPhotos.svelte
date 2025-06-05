<script lang="ts">
	import { createQuery } from '@tanstack/svelte-query';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import * as Card from '$lib/components/ui/card';
	import ImageIcon from '@lucide/svelte/icons/image';
	import ExternalLinkIcon from '@lucide/svelte/icons/external-link';
	import DownloadIcon from '@lucide/svelte/icons/download';
	import { apiGet } from '$lib/utils/api';

	interface Props {
		reportId: number;
	}

	let { reportId }: Props = $props();

	interface PhotoResponse {
		photo_id: number;
		report_id: number;
		filename: string;
		original_filename: string;
		file_size_bytes: number;
		mime_type: string;
		width_pixels?: number;
		height_pixels?: number;
		upload_timestamp: string;
		photo_url: string;
	}

	// Fetch photos for the report
	const photosQuery = $derived(
		createQuery<PhotoResponse[], Error>({
			queryKey: ['reportPhotos', reportId],
			queryFn: async () => {
				return await apiGet<PhotoResponse[]>(`/api/admin/report-photos/${reportId}`);
			},
			enabled: true
		})
	);

	let selectedPhoto = $state<PhotoResponse | null>(null);

	function openPhoto(photo: PhotoResponse) {
		selectedPhoto = photo;
	}

	function closePhotoModal() {
		selectedPhoto = null;
	}

	function formatFileSize(bytes: number): string {
		const units = ['B', 'KB', 'MB', 'GB'];
		let size = bytes;
		let unitIndex = 0;

		while (size >= 1024 && unitIndex < units.length - 1) {
			size /= 1024;
			unitIndex++;
		}

		return `${size.toFixed(1)} ${units[unitIndex]}`;
	}

	function formatUploadTime(timestamp: string): string {
		return new Date(timestamp).toLocaleString();
	}
</script>

{#if $photosQuery.data && $photosQuery.data.length > 0}
	<Card.Root class="p-6">
		<Card.Header class="px-0 pt-0">
			<Card.Title class="flex items-center gap-2">
				<ImageIcon class="h-5 w-5" />
				Photo Evidence
			</Card.Title>
		</Card.Header>
		<Card.Content class="px-0 pb-0">
			<div class="space-y-4">
				<!-- Header -->
				<div class="flex items-center justify-between">
					<div class="flex items-center gap-2">
						<Badge variant="secondary" class="text-xs">
							{$photosQuery.data.length}
							{$photosQuery.data.length === 1 ? 'photo' : 'photos'}
						</Badge>
					</div>
				</div>

				<!-- Photo Grid -->
				<div class="grid grid-cols-2 md:grid-cols-3 gap-4">
					{#each $photosQuery.data as photo (photo.photo_id)}
						<div class="relative group">
							<button
								type="button"
								onclick={() => openPhoto(photo)}
								class="relative block w-full aspect-square rounded-lg overflow-hidden border hover:border-primary/50 transition-colors focus:outline-none focus:ring-2 focus:ring-primary/50"
							>
								<img
									src={photo.photo_url}
									alt={photo.original_filename}
									class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-200"
									loading="lazy"
								/>
								<div
									class="absolute inset-0 bg-black/0 group-hover:bg-black/10 transition-colors duration-200"
								></div>
								<div class="absolute bottom-2 left-2 right-2">
									<div class="bg-black/70 text-white text-xs px-2 py-1 rounded truncate">
										{photo.original_filename}
									</div>
								</div>
							</button>
						</div>
					{/each}
				</div>
			</div>
		</Card.Content>
	</Card.Root>
{/if}

<!-- Photo Modal -->
{#if selectedPhoto}
	<div
		class="fixed inset-0 bg-black/80 z-50 flex items-center justify-center p-4"
		onclick={closePhotoModal}
		role="dialog"
		aria-modal="true"
		aria-label="Photo viewer"
	>
		<div
			class="relative max-w-5xl max-h-full bg-white rounded-lg overflow-hidden"
			onclick={(e) => e.stopPropagation()}
		>
			<!-- Modal Header -->
			<div class="flex items-center justify-between p-4 border-b">
				<div>
					<h3 class="font-medium">{selectedPhoto.original_filename}</h3>
					<div class="flex items-center gap-4 text-sm text-muted-foreground mt-1">
						<span>{formatFileSize(selectedPhoto.file_size_bytes)}</span>
						{#if selectedPhoto.width_pixels && selectedPhoto.height_pixels}
							<span>{selectedPhoto.width_pixels} × {selectedPhoto.height_pixels}</span>
						{/if}
						<span>{formatUploadTime(selectedPhoto.upload_timestamp)}</span>
					</div>
				</div>
				<div class="flex items-center gap-2">
					<Button
						variant="outline"
						size="sm"
						onclick={() => {
							if (selectedPhoto) {
								const link = document.createElement('a');
								link.href = selectedPhoto.photo_url;
								link.download = selectedPhoto.original_filename;
								link.click();
							}
						}}
					>
						<DownloadIcon class="h-4 w-4 mr-1" />
						Download
					</Button>
					<Button
						variant="outline"
						size="sm"
						onclick={() => selectedPhoto && window.open(selectedPhoto.photo_url, '_blank')}
					>
						<ExternalLinkIcon class="h-4 w-4 mr-1" />
						Open
					</Button>
					<Button variant="outline" size="sm" onclick={closePhotoModal}>Close</Button>
				</div>
			</div>

			<!-- Modal Content -->
			<div class="p-4">
				<img
					src={selectedPhoto.photo_url}
					alt={selectedPhoto.original_filename}
					class="max-w-full max-h-[70vh] mx-auto rounded"
				/>
			</div>
		</div>
	</div>
{/if}
