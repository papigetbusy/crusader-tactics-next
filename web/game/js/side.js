// Whose units are whose. In single player the player is side 0. In a versus
// match both sides are players, and the side to move is view.side.
export const isVersus = (v) => !!(v && v.versus);
export const mySide = (v) => (isVersus(v) ? v.side : 0);
// mine is a unit of the side the person holding the phone controls now.
export const mine = (u, v) => !!u && u.team === mySide(v);
