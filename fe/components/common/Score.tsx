export function Score({
  value,
  size = "normal",
}: {
  value: number;
  size?: "normal" | "large";
}) {
  return (
    <div className={`score-pill ${size}`}>
      <span className="score-dot" />
      {value}
      <span className="score-outof">/100</span>
    </div>
  );
}
