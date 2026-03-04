export interface ChatMessage {
    id: string;
    role: "user" | "assistant";
    content: string;
    imageId?: string;
    timestamp: Date;
}

export interface ImagePreview {
    id: string;
    src: string;
    alt: string;
    timestamp: Date;
}
