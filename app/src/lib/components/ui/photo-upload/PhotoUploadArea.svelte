<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import CameraIcon from '@lucide/svelte/icons/camera';
	import ImageIcon from '@lucide/svelte/icons/image';
	import XIcon from '@lucide/svelte/icons/x';

	interface PhotoFile {
		file: File;
		preview: string;
		id: string;
	}

	interface Props {
		photos?: PhotoFile[];
		maxPhotos?: number;
		onPhotosChange?: (photos: PhotoFile[]) => void;
		disabled?: boolean;
		className?: string;
	}

	let {
		photos = $bindable([]),
		maxPhotos = 5,
		onPhotosChange,
		disabled = false,
		className = ''
	}: Props = $props();

	let fileInput: HTMLInputElement;
	let cameraInput: HTMLInputElement;

	// Handle camera capture - should open camera app on mobile
	function handleCameraCapture() {
		cameraInput.click();
	}

	// Handle file selection from gallery
	function handleFileSelect() {
		fileInput.click();
	}

	function handleCameraInputChange(event: Event) {
		const target = event.target as HTMLInputElement;
		const files = target.files;

		if (files) {
			Array.from(files).forEach((file) => {
				if (file.type.startsWith('image/')) {
					addPhoto(file);
				}
			});
		}

		// Reset input
		target.value = '';
	}

	function handleFileInputChange(event: Event) {
		const target = event.target as HTMLInputElement;
		const files = target.files;

		if (files) {
			Array.from(files).forEach((file) => {
				if (file.type.startsWith('image/')) {
					addPhoto(file);
				}
			});
		}

		// Reset input
		target.value = '';
	}

	function addPhoto(file: File) {
		if (photos.length >= maxPhotos) {
			alert(`Maximum ${maxPhotos} photos allowed`);
			return;
		}

		const preview = URL.createObjectURL(file);
		const id = Math.random().toString(36).substr(2, 9);

		const newPhoto: PhotoFile = { file, preview, id };
		photos = [...photos, newPhoto];
		onPhotosChange?.(photos);
	}

	function removePhoto(id: string) {
		const photoToRemove = photos.find((p) => p.id === id);
		if (photoToRemove) {
			URL.revokeObjectURL(photoToRemove.preview);
		}
		photos = photos.filter((p) => p.id !== id);
		onPhotosChange?.(photos);
	}

	// Cleanup on destroy
	function cleanup() {
		photos.forEach((photo) => URL.revokeObjectURL(photo.preview));
	}

	// Call cleanup when component is destroyed
	$effect(() => {
		return cleanup;
	});
</script>

<div class="space-y-3 {className}">
	<!-- Upload buttons -->
	<div class="flex gap-2">
		<Button
			type="button"
			variant="outline"
			size="sm"
			onclick={handleCameraCapture}
			disabled={disabled || photos.length >= maxPhotos}
			class="flex items-center gap-2 flex-1"
		>
			<CameraIcon class="h-4 w-4" />
			<span>Camera</span>
		</Button>

		<Button
			type="button"
			variant="outline"
			size="sm"
			onclick={handleFileSelect}
			disabled={disabled || photos.length >= maxPhotos}
			class="flex items-center gap-2 flex-1"
		>
			<ImageIcon class="h-4 w-4" />
			<span>Gallery</span>
		</Button>

		{#if photos.length > 0}
			<span class="text-sm text-muted-foreground self-center">
				{photos.length}/{maxPhotos} photos
			</span>
		{/if}
	</div>

	<!-- Photo previews -->
	{#if photos.length > 0}
		<div class="grid grid-cols-3 gap-2">
			{#each photos as photo (photo.id)}
				<div class="relative group">
					<img src={photo.preview} alt="" class="w-full h-20 object-cover rounded-lg border" />
					<button
						type="button"
						onclick={() => removePhoto(photo.id)}
						class="absolute -top-1 -right-1 bg-red-500 text-white rounded-full p-1 opacity-0 group-hover:opacity-100 transition-opacity"
						{disabled}
					>
						<XIcon class="h-3 w-3" />
					</button>
					<div class="absolute bottom-1 left-1 bg-black/70 text-white text-xs px-1 rounded">
						{Math.round(photo.file.size / 1024)}KB
					</div>
				</div>
			{/each}
		</div>
	{/if}

	<!-- Camera input - opens camera app on mobile with capture -->
	<input
		type="file"
		bind:this={cameraInput}
		onchange={handleCameraInputChange}
		accept="image/*"
		capture="environment"
		class="hidden"
	/>

	<!-- Gallery input - opens photo library without capture -->
	<input
		type="file"
		bind:this={fileInput}
		onchange={handleFileInputChange}
		accept="image/*"
		multiple
		class="hidden"
	/>
</div>
