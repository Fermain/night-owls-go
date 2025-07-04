<div class="flex flex-wrap gap-6 text-xs text-muted-foreground mt-4 p-4 bg-muted/20 rounded-lg">
	<!-- Shift status indicators -->
	<div class="flex items-center gap-2">
		<div
			class="w-4 h-4 bg-green-500 border border-green-600 rounded flex items-center justify-center shadow-sm"
		>
			<span class="text-[8px] font-bold text-white">✓</span>
		</div>
		<span>Filled shift (assigned)</span>
	</div>
	<div class="flex items-center gap-2">
		<div
			class="w-4 h-4 bg-red-500 border border-red-600 rounded flex items-center justify-center shadow-sm"
		>
			<span class="text-[8px] font-bold text-white">!</span>
		</div>
		<span>Unfilled shift (needs assignment)</span>
	</div>
	<div class="flex items-center gap-2">
		<div
			class="w-4 h-4 bg-yellow-500 border border-yellow-600 rounded flex items-center justify-center shadow-sm"
		>
			<span class="text-[8px] font-bold text-white">?</span>
		</div>
		<span>Data inconsistency</span>
	</div>

	<div class="w-px h-4 bg-muted-foreground/30"></div>

	<!-- Calendar state indicators -->
	<div class="flex items-center gap-2">
		<div class="w-4 h-4 border-2 border-primary rounded"></div>
		<span>Today</span>
	</div>
	<div class="flex items-center gap-2">
		<div class="w-4 h-4 bg-gray-400 border border-gray-500 rounded opacity-50"></div>
		<span>Past date</span>
	</div>
</div>
