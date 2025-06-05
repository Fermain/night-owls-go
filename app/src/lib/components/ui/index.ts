/**
 * UI Components Index
 * Centralized exports for all reusable UI components
 */

// === STATE COMPONENTS ===
export { LoadingState } from './loading-state';
export { ErrorState } from './error-state';
export { EmptyState } from './empty-state';

// === CORE COMPONENTS ===
export { Button } from './button';
export { Input } from './input';
export { Label } from './label';
export { Textarea } from './textarea';
export { Select, SelectContent, SelectItem, SelectTrigger } from './select';
export { Checkbox } from './checkbox';
export { Switch } from './switch';
export { Progress } from './progress';

// === LAYOUT COMPONENTS ===
export { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from './card';
export { Separator } from './separator';
export { Badge } from './badge';
export { Avatar, AvatarFallback, AvatarImage } from './avatar';

// === NAVIGATION COMPONENTS ===
export {
	Breadcrumb,
	BreadcrumbEllipsis,
	BreadcrumbItem,
	BreadcrumbLink,
	BreadcrumbList,
	BreadcrumbPage,
	BreadcrumbSeparator
} from './breadcrumb';

// === OVERLAY COMPONENTS ===
export {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
	DialogTrigger
} from './dialog';
export {
	AlertDialog,
	AlertDialogAction,
	AlertDialogCancel,
	AlertDialogContent,
	AlertDialogDescription,
	AlertDialogFooter,
	AlertDialogHeader,
	AlertDialogTitle,
	AlertDialogTrigger
} from './alert-dialog';
export {
	Sheet,
	SheetContent,
	SheetDescription,
	SheetFooter,
	SheetHeader,
	SheetTitle,
	SheetTrigger
} from './sheet';
export { Popover, PopoverContent, PopoverTrigger } from './popover';
export { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from './tooltip';

// === SPECIALIZED COMPONENTS ===
export { Calendar } from './calendar';
export {
	Table,
	TableBody,
	TableCaption,
	TableCell,
	TableFooter,
	TableHead,
	TableHeader,
	TableRow
} from './table';
export {
	Pagination,
	PaginationContent,
	PaginationEllipsis,
	PaginationItem,
	PaginationLink
} from './pagination';

// === APP-SPECIFIC COMPONENTS ===
export { PhoneInput } from './phone-input';
export { UserMultiSelect } from './user-multiselect';

// === UTILITY COMPONENTS ===
export { Skeleton } from './skeleton';
export {
	Command,
	CommandDialog,
	CommandEmpty,
	CommandGroup,
	CommandInput,
	CommandItem,
	CommandList,
	CommandSeparator,
	CommandShortcut
} from './command';
export {
	DropdownMenu,
	DropdownMenuCheckboxItem,
	DropdownMenuContent,
	DropdownMenuGroup,
	DropdownMenuItem,
	DropdownMenuLabel,
	DropdownMenuRadioGroup,
	DropdownMenuRadioItem,
	DropdownMenuSeparator,
	DropdownMenuShortcut,
	DropdownMenuSub,
	DropdownMenuSubContent,
	DropdownMenuSubTrigger,
	DropdownMenuTrigger
} from './dropdown-menu';

// === FORMS ===
export { InputOTP, InputOTPGroup, InputOTPSeparator, InputOTPSlot } from './input-otp';

// === CALENDAR ===
export { RangeCalendar } from './range-calendar';

// === SIDEBAR ===
export {
	Sidebar,
	SidebarContent,
	SidebarFooter,
	SidebarGroup,
	SidebarGroupAction,
	SidebarGroupContent,
	SidebarGroupLabel,
	SidebarHeader,
	SidebarInset,
	SidebarMenu,
	SidebarMenuAction,
	SidebarMenuBadge,
	SidebarMenuButton,
	SidebarMenuItem,
	SidebarMenuSkeleton,
	SidebarMenuSub,
	SidebarMenuSubButton,
	SidebarMenuSubItem,
	SidebarProvider,
	SidebarRail,
	SidebarSeparator,
	SidebarTrigger
} from './sidebar';
