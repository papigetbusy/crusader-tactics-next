// Whose units are whose. In single player the player is side 0. In a versus
// match on one phone, the side to move holds the phone. Online, this phone
// always plays its own seat (or only watches: seat -1).
let seat = null; // online: this phone's side, -1 watching; null when not online
export const setSeat = (s) => { seat = s; };
export const isVersus = (v) => !!(v && v.versus);
export const mySide = (v) => (isVersus(v) ? (seat !== null ? seat : v.side) : 0);
// mine is a unit of the side this phone controls.
export const mine = (u, v) => !!u && u.team === mySide(v);
// myTurn is whether this phone may act now.
export const myTurn = (v) => !isVersus(v) || seat === null || (v.side === seat && !v.outcome);
// onlineSeat is this phone's online seat, or null when not playing online.
export const onlineSeat = () => seat;
