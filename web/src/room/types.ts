export interface ChatMessage {
    id: string;
    role: "user" | "assistant";
    content: string;
    kind?: "default" | "screenshot" | "ai";
    imageId?: string;
    streaming?: boolean;
    fullWidth?: boolean;
    timestamp: Date;
}

export interface ImagePreview {
    id: string;
    src: string;
    alt: string;
    timestamp: Date;
}
