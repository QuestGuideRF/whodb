import {Button} from "@clidey/ux";
import {Component, type ErrorInfo, type ReactNode} from "react";
import {captureException} from "../config/posthog";
import {withBasePath} from "../utils/base-path";
import {openExternalLink} from "../utils/external-links";

interface ErrorBoundaryProps {
    children: ReactNode;
}

interface ErrorBoundaryState {
    hasError: boolean;
}

export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
    state: ErrorBoundaryState = {hasError: false};

    static getDerivedStateFromError(): ErrorBoundaryState {
        return {hasError: true};
    }

    componentDidCatch(error: Error, info: ErrorInfo) {
        console.error("Uncaught error in component tree:", error, info.componentStack);
        captureException(error, {componentStack: info.componentStack ?? ""});
    }

    private handleGoHome = () => {
        window.location.href = withBasePath("/");
    };

    render() {
        if (this.state.hasError) {
            return (
                <div className="flex flex-col items-center justify-center min-h-screen gap-6 p-8">
                    <h1 className="text-3xl font-bold text-foreground">
                        Whoops, something went wrong.
                    </h1>
                    <p className="text-muted-foreground">
                        Please either refresh the page or return home to try again.
                    </p>
                    <p className="text-muted-foreground">
                        If the issue continues, please{" "}
                        <a
                            href="https://github.com/clidey/whodb/issues"
                            onClick={(e) => openExternalLink("https://github.com/clidey/whodb/issues", e)}
                            className="text-primary underline underline-offset-2"
                        >
                            get in touch.
                        </a>
                    </p>
                    <Button onClick={this.handleGoHome}>Go home</Button>
                </div>
            );
        }
        return this.props.children;
    }
}