import { writable } from "svelte/store";

/** @type {import("svelte/store").Writable<Array<{ id: number; type: 'success'|'error'|'warning'; message: string }>>} */
export const notifications = writable([]);

let nextId = 1;

/**
 * @param {'success'|'error'|'warning'} type
 * @param {string} message
 */
export function addNotification(type, message) {
  const id = nextId++;
  notifications.update((list) => [...list, { id, type, message }]);
  return id;
}

/** @param {number} id */
export function dismissNotification(id) {
  notifications.update((list) => list.filter((n) => n.id !== id));
}
