export function fmtSecond(seconds, fmt="{d}天 {h}时 {m}分 {s}秒") {
  const d = Math.floor(seconds / 86400);
  const h = Math.floor((seconds % 86400) / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const s = seconds % 60;
  return fmt.replace("{d}", d).replace("{h}", h).replace("{m}", m).replace("{s}", s);
}