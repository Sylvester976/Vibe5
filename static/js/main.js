// Smooth scrolling for navigation links
document.querySelectorAll('a[href^="#"]').forEach(anchor => {
    anchor.addEventListener('click', function (e) {
        e.preventDefault();
        const target = document.querySelector(this.getAttribute('href'));
        if (target) {
            target.scrollIntoView({
                behavior: 'smooth',
                block: 'start'
            });
        }
    });
});

// Add scroll effect to navbar
window.addEventListener('scroll', function() {
    const navbar = document.querySelector('.navbar');
    if (window.scrollY > 100) {
        navbar.style.background = 'rgba(25, 20, 20, 0.95)';
    } else {
        navbar.style.background = 'rgba(25, 20, 20, 0.9)';
    }
});

// Animate stats on scroll
const animateStats = () => {
    const stats = document.querySelectorAll('.stat-number');
    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                const target = entry.target;
                const finalValue = target.textContent;
                target.textContent = '0';

                if (finalValue.includes('K')) {
                    animateNumber(target, parseInt(finalValue) * 1000, finalValue);
                } else if (finalValue.includes('M')) {
                    animateNumber(target, parseInt(finalValue) * 1000000, finalValue);
                } else {
                    target.textContent = finalValue;
                }
            }
        });
    });

    stats.forEach(stat => observer.observe(stat));
};

const animateNumber = (element, targetValue, originalText) => {
    const duration = 2000;
    const increment = targetValue / (duration / 16);
    let currentValue = 0;

    const timer = setInterval(() => {
        currentValue += increment;
        if (currentValue >= targetValue) {
            element.textContent = originalText;
            clearInterval(timer);
        } else {
            if (originalText.includes('K')) {
                element.textContent = Math.floor(currentValue / 1000) + 'K+';
            } else if (originalText.includes('M')) {
                element.textContent = Math.floor(currentValue / 1000000) + 'M+';
            }
        }
    }, 16);
};

// Initialize animations
animateStats();