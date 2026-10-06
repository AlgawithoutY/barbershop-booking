package policies

// Definisikan konstanta untuk Role pengguna
const (
	RoleAdmin    = "admin"
	RoleCustomer = "customer"
)

// Definisikan policy/permission khusus jika diperlukan di masa depan
const (
	CanManageBarbers = "manage_barbers"
	CanManageServices = "manage_services"
	CanBookSlot      = "book_slot"
)