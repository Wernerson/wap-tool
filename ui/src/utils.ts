export function parseTime(time: string) : Date {
    const today = new Date().toISOString().split('T')[0];
    return new Date(`${today}T${time}:00`);
}