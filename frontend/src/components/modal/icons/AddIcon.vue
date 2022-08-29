<template>
  <div class="icon">
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
      <path
        d="M17,13H13V17H11V13H7V11H11V7H13V11H17M12,2A10,10 0 0,0 2,12A10,10 0 0,0 12,22A10,10 0 0,0 22,12A10,10 0 0,0 12,2Z"
      />
    </svg>
  </div>
</template>

<style scoped>
.icon {
  opacity: 0.6;
  transition: opacity 0.15s ease;
}

.icon:hover {
  cursor: pointer;
  opacity: 1;
}

svg {
  width: 24px;
  height: 24px;
  fill: #64bcff;
  transition: fill 200ms ease;
  margin-left: 10px;
  margin-right: 5px;
}

.icon:hover svg {
  fill: #defefc;
}
</style>
