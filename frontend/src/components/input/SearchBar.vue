<template>
  <input class="search-input" placeholder="SEARCH" />
</template>

<style scoped>
input::-webkit-input-placeholder {
  color: hsla(0, 0%, 100%, 0.3);
}

input::-moz-placeholder {
  color: hsla(0, 0%, 100%, 0.3);
}

input::placeholder {
  color: hsla(0, 0%, 100%, 0.3);
}

input,
select,
textarea {
  font: inherit;
  line-height: 1.5;
  border: none;
  display: block;
  box-sizing: border-box;
  width: 100%;
  padding: 8px 16px;
  border-radius: 2px;
  resize: none;
}

.search-input {
  color: hsla(0, 0%, 100%, 0.8);
  background: #194f85;
  top: 0;
  margin-bottom: 0 !important;
}
</style>
