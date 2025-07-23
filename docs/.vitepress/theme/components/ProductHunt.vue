<template>
  <div class="product-hunt-badge">
    <a href="https://www.producthunt.com/products/ccproxy-any-ai-model-with-claude-code?embed=true&utm_source=badge-featured&utm_medium=badge&utm_source=badge-ccproxy" 
       target="_blank"
       rel="noopener noreferrer">
      <img src="https://api.producthunt.com/widgets/embed-image/v1/featured.svg?post_id=996378&theme=light&t=1753258354865" 
           alt="CCProxy - Keep Claude Code's magic, slash costs by 90% | Product Hunt" 
           style="width: 250px; height: 54px;" 
           width="250" 
           height="54" />
    </a>
  </div>
</template>

<style scoped>
.product-hunt-badge {
  display: flex;
  justify-content: center;
  align-items: center;
  margin: 0;
  padding: 12px 0;
}

.product-hunt-badge a {
  display: inline-block;
  transition: transform 0.2s ease;
}

.product-hunt-badge a:hover {
  transform: translateY(-2px);
}

/* Dark mode support */
.dark .product-hunt-badge {
  filter: brightness(0.9);
}

/* Mobile responsive */
@media (max-width: 640px) {
  .product-hunt-badge img {
    width: 200px !important;
    height: 43px !important;
  }
}
</style>