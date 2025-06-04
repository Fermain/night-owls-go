// Smart polling configuration for different data types in community watch app

export const queryConfig = {
	// Critical data - refetch more frequently
	critical: {
		refetchInterval: 30 * 1000, // 30 seconds
		staleTime: 1 * 60 * 1000, // 1 minute
		gcTime: 5 * 60 * 1000 // 5 minutes
	},

	// Important data - moderate polling
	important: {
		refetchInterval: 2 * 60 * 1000, // 2 minutes
		staleTime: 5 * 60 * 1000, // 5 minutes
		gcTime: 10 * 60 * 1000 // 10 minutes
	},

	// Normal data - standard polling
	normal: {
		refetchInterval: 5 * 60 * 1000, // 5 minutes
		staleTime: 10 * 60 * 1000, // 10 minutes
		gcTime: 30 * 60 * 1000 // 30 minutes
	},

	// Static data - rarely changes
	static: {
		refetchInterval: false, // No automatic refetch
		staleTime: 30 * 60 * 1000, // 30 minutes
		gcTime: 60 * 60 * 1000 // 1 hour
	}
};

// Data type mappings for community watch
export const dataTypeConfig = {
	// Critical: emergency contacts, active reports
	emergencyContacts: queryConfig.critical,
	activeReports: queryConfig.critical,
	broadcasts: queryConfig.critical,

	// Important: shifts, user data, analytics
	shifts: queryConfig.important,
	userBookings: queryConfig.important,
	reports: queryConfig.important,

	// Normal: general admin data
	users: queryConfig.normal,
	schedules: queryConfig.normal,
	history: queryConfig.normal,

	// Static: rarely changing data
	systemSettings: queryConfig.static
};

// Helper to get config for a data type
export function getQueryConfig(dataType: keyof typeof dataTypeConfig) {
	return dataTypeConfig[dataType] || queryConfig.normal;
}
