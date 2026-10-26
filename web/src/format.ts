export const money = new Intl.NumberFormat("en-US", { style: "currency", currency: "USD" });

export const percent = (part: number, whole: number) => (whole === 0 ? "-" : `${((part / whole) * 100).toFixed(1)}%`);
