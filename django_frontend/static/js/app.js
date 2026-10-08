document.addEventListener("DOMContentLoaded", () => {
    const sidebar = document.getElementById("sidebar");
    const menuButton = document.getElementById("menuButton");
    const closeButton = document.getElementById("sidebarClose");
    const overlay = document.getElementById("sidebarOverlay");

    const openSidebar = () => {
        sidebar?.classList.add("open");
        overlay?.classList.add("show");
        document.body.style.overflow = "hidden";
    };

    const closeSidebar = () => {
        sidebar?.classList.remove("open");
        overlay?.classList.remove("show");
        document.body.style.overflow = "";
    };

    menuButton?.addEventListener("click", openSidebar);
    closeButton?.addEventListener("click", closeSidebar);
    overlay?.addEventListener("click", closeSidebar);

    window.addEventListener("keydown", (event) => {
        if (event.key === "Escape") {
            closeSidebar();
        }
    });
});