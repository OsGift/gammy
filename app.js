/* app.js */
/* LocalStorage demo logic + new enhancements */

// ======================== GLOBAL STATE ========================
let currentUser = null;
let users = JSON.parse(localStorage.getItem("users")) || [];
let bookings = JSON.parse(localStorage.getItem("bookings")) || [];

// ======================== DOM ELEMENTS ========================
// Landing
const landingPage = document.getElementById("landing-page");
const signupSection = document.getElementById("signup-section");
const loginSection = document.getElementById("login-section");
const bookingSection = document.getElementById("booking-section");
const providerDashboard = document.getElementById("provider-dashboard");
const adminDashboard = document.getElementById("admin-dashboard");

// Forms
const signupForm = document.getElementById("signup-form");
const loginForm = document.getElementById("login-form");
const bookingForm = document.getElementById("booking-form");
const profileUpdateForm = document.getElementById("profile-update-form");

// Nav
const navbar = document.getElementById("navbar");
const navDashboard = document.getElementById("nav-dashboard");
const navProfile = document.getElementById("nav-profile");
const navLogout = document.getElementById("nav-logout");

// Profile modal
const profileModal = document.getElementById("profile-modal");
const closeProfileModal = document.getElementById("close-profile-modal");

// Booking success modal
const bookingSuccessModal = document.getElementById("booking-success-modal");
const closeBookingSuccess = document.getElementById("close-booking-success");

// NEW FEATURE START: Provider availability toggle
const toggleAvailabilityBtn = document.getElementById("toggle-availability-btn");
const providerApprovalStatus = document.getElementById("provider-approval-status");
// NEW FEATURE END

// ======================== EVENT LISTENERS ========================

// Show signup
document.getElementById("show-signup").addEventListener("click", () => {
    landingPage.classList.add("hidden");
    signupSection.classList.remove("hidden");
});

// Show login
document.getElementById("show-login").addEventListener("click", () => {
    landingPage.classList.add("hidden");
    loginSection.classList.remove("hidden");
});

document.getElementById("link-to-login").addEventListener("click", () => {
    signupSection.classList.add("hidden");
    loginSection.classList.remove("hidden");
});

document.getElementById("link-to-signup").addEventListener("click", () => {
    loginSection.classList.add("hidden");
    signupSection.classList.remove("hidden");
});

// Role change in signup — show services if provider
document.getElementById("signup-role").addEventListener("change", (e) => {
    const container = document.getElementById("services-offered-container");
    if (e.target.value === "mower") {
        container.classList.remove("hidden");
    } else {
        container.classList.add("hidden");
    }
});

// Signup
signupForm.addEventListener("submit", (e) => {
    e.preventDefault();
    const name = document.getElementById("signup-name").value.trim();
    const email = document.getElementById("signup-email").value.trim();
    const password = document.getElementById("signup-password").value;
    const role = document.getElementById("signup-role").value;
    let services = "";

    // NEW FEATURE: Save services for providers
    if (role === "mower") {
        services = document.getElementById("services").value.trim();
    }

    if (users.some(u => u.email === email)) {
        alert("Email already registered.");
        return;
    }

    const newUser = {
        id: Date.now(),
        name,
        email,
        password,
        role,
        services,
        isApproved: role === "mower" ? false : true, // Providers need admin approval
        isAvailable: false,
        completedBookings: 0,
        rating: 0,
        profilePicture: "",
    };

    users.push(newUser);
    localStorage.setItem("users", JSON.stringify(users));
    alert("Signup successful. Please log in.");
    signupSection.classList.add("hidden");
    loginSection.classList.remove("hidden");
});

// Login
loginForm.addEventListener("submit", (e) => {
    e.preventDefault();
    const email = document.getElementById("login-email").value.trim();
    const password = document.getElementById("login-password").value;

    const user = users.find(u => u.email === email && u.password === password);
    if (!user) {
        alert("Invalid credentials.");
        return;
    }

    currentUser = user;
    localStorage.setItem("currentUser", JSON.stringify(currentUser));

    loginSection.classList.add("hidden");
    navbar.classList.remove("hidden");
    showDashboard();
});

// Logout
navLogout.addEventListener("click", () => {
    currentUser = null;
    localStorage.removeItem("currentUser");
    navbar.classList.add("hidden");
    landingPage.classList.remove("hidden");
    hideAllSections();
});

// Show dashboard based on role
function showDashboard() {
    hideAllSections();
    if (currentUser.role === "customer") {
        bookingSection.classList.remove("hidden");
        renderAvailableProviders();
    } else if (currentUser.role === "mower") {
        providerDashboard.classList.remove("hidden");
        updateProviderDashboard();
    } else if (currentUser.role === "admin") {
        adminDashboard.classList.remove("hidden");
        updateAdminDashboard();
    }
}

function hideAllSections() {
    bookingSection.classList.add("hidden");
    providerDashboard.classList.add("hidden");
    adminDashboard.classList.add("hidden");
}

// ======================== CUSTOMER BOOKING ========================
function renderAvailableProviders() {
    const container = document.getElementById("available-providers-list");
    container.innerHTML = "";
    const providers = users.filter(u => u.role === "mower" && u.isApproved && u.isAvailable);
    if (providers.length === 0) {
        container.innerHTML = "<p>No available providers at the moment.</p>";
        return;
    }
    providers.forEach(provider => {
        const div = document.createElement("div");
        div.classList.add("provider-card");
        div.innerHTML = `
            <h4>${provider.name}</h4>
            <p>Services: ${provider.services || "N/A"}</p>
            <p>Completed Bookings: ${provider.completedBookings}</p>
            <p>Rating: ${provider.rating} ⭐</p>
            <button onclick="selectProvider(${provider.id})">Select</button>
        `;
        container.appendChild(div);
    });
}

let selectedProviderId = null;
window.selectProvider = function (id) {
    selectedProviderId = id;
    alert("Provider selected.");
};

bookingForm.addEventListener("submit", (e) => {
    e.preventDefault();
    if (!selectedProviderId) {
        alert("Please select a provider.");
        return;
    }
    const date = document.getElementById("booking-date").value;
    const description = document.getElementById("booking-description").value;

    const newBooking = {
        id: Date.now(),
        customerId: currentUser.id,
        providerId: selectedProviderId,
        date,
        description,
        status: "Pending",
    };

    bookings.push(newBooking);
    localStorage.setItem("bookings", JSON.stringify(bookings));

    bookingSuccessModal.classList.remove("hidden");
});

closeBookingSuccess.addEventListener("click", () => {
    bookingSuccessModal.classList.add("hidden");
});

// ======================== PROVIDER DASHBOARD ========================
function updateProviderDashboard() {
    // Show approval status
    if (!currentUser.isApproved) {
        providerApprovalStatus.textContent = "Your account is pending admin approval.";
        toggleAvailabilityBtn.classList.add("hidden");
    } else {
        providerApprovalStatus.textContent = "";
        toggleAvailabilityBtn.classList.remove("hidden");
        toggleAvailabilityBtn.textContent = currentUser.isAvailable ? "Set as Unavailable" : "Set as Available";
    }

    // Render provider bookings
    const tbody = document.querySelector("#provider-bookings-table tbody");
    tbody.innerHTML = "";
    const myBookings = bookings.filter(b => b.providerId === currentUser.id);
    myBookings.forEach(b => {
        const customer = users.find(u => u.id === b.customerId);
        const tr = document.createElement("tr");
        tr.innerHTML = `
            <td class="border px-2 py-1">${b.id}</td>
            <td class="border px-2 py-1">${customer ? customer.name : "Unknown"}</td>
            <td class="border px-2 py-1">${b.status}</td>
            <td class="border px-2 py-1">
                ${b.status !== "Completed" ? `<button onclick="markBookingCompleted(${b.id})" class="px-2 py-1 bg-green-500 text-white rounded text-xs">Mark Completed</button>` : ""}
            </td>
        `;
        tbody.appendChild(tr);
    });
}

toggleAvailabilityBtn.addEventListener("click", () => {
    currentUser.isAvailable = !currentUser.isAvailable;
    users = users.map(u => u.id === currentUser.id ? currentUser : u);
    localStorage.setItem("users", JSON.stringify(users));
    updateProviderDashboard();
});

window.markBookingCompleted = function (bookingId) {
    bookings = bookings.map(b => {
        if (b.id === bookingId) {
            b.status = "Completed";
            const provider = users.find(u => u.id === b.providerId);
            if (provider) {
                provider.completedBookings++;
                localStorage.setItem("users", JSON.stringify(users));
            }
        }
        return b;
    });
    localStorage.setItem("bookings", JSON.stringify(bookings));
    updateProviderDashboard();
};

// ======================== ADMIN DASHBOARD ========================
function updateAdminDashboard() {
    // Analytics
    document.getElementById("total-users-count").textContent = users.length;
    document.getElementById("total-providers-count").textContent = users.filter(u => u.role === "mower").length;
    document.getElementById("pending-approvals-count").textContent = users.filter(u => u.role === "mower" && !u.isApproved).length;
    document.getElementById("total-bookings-count").textContent = bookings.length;

    // Providers table
    const providersTbody = document.querySelector("#admin-providers-table tbody");
    providersTbody.innerHTML = "";
    users.filter(u => u.role === "mower").forEach(provider => {
        const tr = document.createElement("tr");
        tr.innerHTML = `
            <td class="border px-2 py-1">${provider.name}</td>
            <td class="border px-2 py-1">${provider.services || "N/A"}</td>
            <td class="border px-2 py-1">${provider.isApproved ? "Approved" : "Pending"}</td>
            <td class="border px-2 py-1">${provider.isAvailable ? "Available" : "Unavailable"}</td>
            <td class="border px-2 py-1">
                ${!provider.isApproved ? `<button onclick="approveProvider(${provider.id})" class="bg-green-500 text-white px-2 py-1 rounded text-xs">Approve</button>` : ""}
                ${provider.isApproved ? `<button onclick="toggleProviderAvailability(${provider.id})" class="bg-blue-500 text-white px-2 py-1 rounded text-xs">Toggle Availability</button>` : ""}
                <button onclick="declineProvider(${provider.id})" class="bg-red-500 text-white px-2 py-1 rounded text-xs">Decline</button>
            </td>
        `;
        providersTbody.appendChild(tr);
    });

    // Customers table
    const customersTbody = document.querySelector("#admin-customers-table tbody");
    customersTbody.innerHTML = "";
    users.filter(u => u.role === "customer").forEach(customer => {
        const tr = document.createElement("tr");
        tr.innerHTML = `
            <td class="border px-2 py-1">${customer.name}</td>
            <td class="border px-2 py-1">${customer.email}</td>
        `;
        customersTbody.appendChild(tr);
    });

    // Admins table
    const adminsTbody = document.querySelector("#admin-admins-table tbody");
    adminsTbody.innerHTML = "";
    users.filter(u => u.role === "admin").forEach(admin => {
        const tr = document.createElement("tr");
        tr.innerHTML = `
            <td class="border px-2 py-1">${admin.name}</td>
            <td class="border px-2 py-1">${admin.email}</td>
        `;
        adminsTbody.appendChild(tr);
    });
}

window.approveProvider = function (id) {
    users = users.map(u => u.id === id ? { ...u, isApproved: true } : u);
    localStorage.setItem("users", JSON.stringify(users));
    updateAdminDashboard();
};

window.declineProvider = function (id) {
    users = users.filter(u => u.id !== id);
    localStorage.setItem("users", JSON.stringify(users));
    updateAdminDashboard();
};

window.toggleProviderAvailability = function (id) {
    users = users.map(u => u.id === id ? { ...u, isAvailable: !u.isAvailable } : u);
    localStorage.setItem("users", JSON.stringify(users));
    updateAdminDashboard();
};

// ======================== PROFILE UPDATE ========================
navProfile.addEventListener("click", () => {
    document.getElementById("profile-name").value = currentUser.name;
    document.getElementById("profile-email").value = currentUser.email;
    if (currentUser.role === "mower") {
        document.getElementById("profile-services-container").classList.remove("hidden");
        document.getElementById("profile-services").value = currentUser.services || "";
    } else {
        document.getElementById("profile-services-container").classList.add("hidden");
    }
    profileModal.classList.remove("hidden");
});

closeProfileModal.addEventListener("click", () => {
    profileModal.classList.add("hidden");
});

profileUpdateForm.addEventListener("submit", (e) => {
    e.preventDefault();
    currentUser.name = document.getElementById("profile-name").value.trim();
    currentUser.email = document.getElementById("profile-email").value.trim();
    if (currentUser.role === "mower") {
        currentUser.services = document.getElementById("profile-services").value.trim();
    }
    const newPassword = document.getElementById("profile-password").value.trim();
    if (newPassword) {
        currentUser.password = newPassword;
    }
    // Note: Profile picture file handling would need server-side storage in a real app
    users = users.map(u => u.id === currentUser.id ? currentUser : u);
    localStorage.setItem("users", JSON.stringify(users));
    profileModal.classList.add("hidden");
    showDashboard();
});

// ======================== RESTORE SESSION ========================
window.addEventListener("load", () => {
    const savedUser = localStorage.getItem("currentUser");
    if (savedUser) {
        currentUser = JSON.parse(savedUser);
        navbar.classList.remove("hidden");
        showDashboard();
    }
});
