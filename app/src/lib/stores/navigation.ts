import { readable, derived } from 'svelte/store';
import { userSession } from './authStore';
import Inbox from '@lucide/svelte/icons/inbox';
import Users from '@lucide/svelte/icons/users';
import Send from '@lucide/svelte/icons/send';
import ListChecks from '@lucide/svelte/icons/list-checks';
import Home from '@lucide/svelte/icons/home';
import Calendar from '@lucide/svelte/icons/calendar';
import MessageCircle from '@lucide/svelte/icons/message-circle';
import AlertTriangle from '@lucide/svelte/icons/alert-triangle';
import ClipboardList from '@lucide/svelte/icons/clipboard-list';

export interface NavItem {
	title: string;
	url: string;
	icon: typeof Inbox;
	description?: string;
	roles?: ('admin' | 'owl' | 'guest')[];
}

// Admin navigation items
const adminNavItems: NavItem[] = [
	{
		title: 'Reports',
		url: '/admin/reports',
		icon: Inbox,
		description: 'View incident reports',
		roles: ['admin']
	},
	{
		title: 'Shifts',
		url: '/admin/shifts',
		icon: ListChecks,
		description: 'Manage shift schedules',
		roles: ['admin']
	},
	{
		title: 'Users',
		url: '/admin/users',
		icon: Users,
		description: 'Manage community members',
		roles: ['admin']
	},
	{
		title: 'Broadcasts',
		url: '/admin/broadcasts',
		icon: Send,
		description: 'Send community messages',
		roles: ['admin']
	}
];

// Public navigation items
const publicNavItems: NavItem[] = [
	{
		title: 'Home',
		url: '/',
		icon: Home,
		description: 'Community dashboard',
		roles: ['admin', 'owl', 'guest']
	},
	{
		title: 'Shifts',
		url: '/shifts',
		icon: Calendar,
		description: 'Available patrol shifts',
		roles: ['admin', 'owl']
	},
	{
		title: 'Messages',
		url: '/broadcasts',
		icon: MessageCircle,
		description: 'Community announcements',
		roles: ['admin', 'owl']
	},
	{
		title: 'Report',
		url: '/report',
		icon: AlertTriangle,
		description: 'Report an incident',
		roles: ['admin', 'owl', 'guest']
	},
	{
		title: 'My Bookings',
		url: '/bookings',
		icon: ClipboardList,
		description: 'Your shift bookings',
		roles: ['admin', 'owl']
	}
];

// Legacy admin navigation (for backward compatibility)
export const navigation = readable(adminNavItems);

// Role-based navigation
export const adminNavigation = readable(adminNavItems);
export const publicNavigation = readable(publicNavItems);

// Dynamic navigation based on user role and context
export const contextualNavigation = derived(
	[userSession],
	([$userSession]) => {
		const userRole = $userSession.role || 'guest';
		
		// Filter items based on user role
		const filterByRole = (items: NavItem[]) => 
			items.filter(item => {
				if (!item.roles) return true;
				return item.roles.includes(userRole as 'admin' | 'owl' | 'guest');
			});

		return {
			admin: filterByRole(adminNavItems),
			public: filterByRole(publicNavItems),
			userRole,
			isAuthenticated: $userSession.isAuthenticated
		};
	}
);
