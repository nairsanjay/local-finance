const header = document.querySelector('.site-header');
const revealItems = document.querySelectorAll('.statement, .workflow-step, .section-heading, .feature-index article, .privacy-copy, .architecture, .closing > *');

const observer = new IntersectionObserver((entries) => {
  entries.forEach((entry) => {
    if (entry.isIntersecting) {
      entry.target.classList.add('is-visible');
      observer.unobserve(entry.target);
    }
  });
}, { threshold: 0.14 });

revealItems.forEach((item) => {
  item.classList.add('reveal');
  observer.observe(item);
});

window.addEventListener('scroll', () => {
  header.classList.toggle('is-scrolled', window.scrollY > 24);
}, { passive: true });
