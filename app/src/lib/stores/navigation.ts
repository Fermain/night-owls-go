import { readable } from 'svelte/store';
import Inbox from '@lucide/svelte/icons/inbox';
import CalendarRange from '@lucide/svelte/icons/calendar-range';
import ChartCandlestick from '@lucide/svelte/icons/chart-candlestick';
import Users from '@lucide/svelte/icons/users';
import Send from '@lucide/svelte/icons/send';
import ListChecks from '@lucide/svelte/icons/list-checks';

interface NavItem {
	title: string;
	url: string;
	icon: typeof Inbox;
}

const navMain: NavItem[] = [
	{
		title: 'Reports',
		url: '/admin/reports',
		icon: Inbox
	},
	{
		title: 'Shifts',
		url: '/admin/schedules',
		icon: CalendarRange
	},
	{
		title: 'Shift Slots',
		url: '/admin/schedules/slots',
		icon: ListChecks
	},
	{
		title: 'Statistics',
		url: '#',
		icon: ChartCandlestick
	},
	{
		title: 'Users',
		url: '/admin/users',
		icon: Users
	},
	{
		title: 'Broadcasts',
		url: '/admin/broadcasts',
		icon: Send
	}
];

export const navigation = readable(navMain);
