import { writable } from "svelte/store";

const savedToken = localStorage.getItem("go_nmos_token");
const savedUser = localStorage.getItem("go_nmos_user");

export const auth = writable({
  token: savedToken,
  user: savedUser ? JSON.parse(savedUser) : null
});

export function setAuth(token, user) {
  localStorage.setItem("go_nmos_token", token);
  localStorage.setItem("go_nmos_user", JSON.stringify(user));
  auth.set({ token, user });
}

export function clearAuth() {
  localStorage.removeItem("go_nmos_token");
  localStorage.removeItem("go_nmos_user");
  auth.set({ token: null, user: null });
}
