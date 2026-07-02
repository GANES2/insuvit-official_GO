document.addEventListener('DOMContentLoaded', () => {
    // Scroll Spy - highlight the nav link of the section currently in view
    const navMap = {};
    document.querySelectorAll('.nav-menu a[href^="#"]').forEach(link => {
        navMap[link.getAttribute('href').slice(1)] = link;
    });

    const navbar = document.getElementById('navbar');

    function updateActiveNav() {
        if (!navbar) return;
        const scrollPos = window.scrollY + navbar.offsetHeight + 40;
        let currentId = '';
        document.querySelectorAll('section[id]').forEach(section => {
            if (section.offsetTop <= scrollPos) {
                currentId = section.id;
            }
        });
        
        Object.values(navMap).forEach(link => link.classList.remove('active'));
        if (currentId && navMap[currentId]) {
            navMap[currentId].classList.add('active');
        }
    }

    window.addEventListener('scroll', updateActiveNav);
    updateActiveNav();
});
