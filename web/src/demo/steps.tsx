// One placeholder per demo step. Each will be replaced by the real step as it is built.

const placeholder = (text: string) => <p className="notice">{text}</p>;

export const UploadStep = () => placeholder("Upload: choose the sample statement. (not built yet)");
export const ProcessStep = () => placeholder("Process: crop, OCR and extraction, with a view of each stage. (not built yet)");
export const ReviewStep = () => placeholder("Review: confirm the rows JEV was unsure about. (not built yet)");
export const DashboardStep = () => placeholder("Dashboard: spending by month and category, essential split. (not built yet)");
export const EditStep = () => placeholder("Edit data: change a category or essential flag and the dashboard updates. (not built yet)");
