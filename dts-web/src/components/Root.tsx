import { Theme } from "@radix-ui/themes";
import { useNavigate } from "react-router-dom";
import { AuthKitProvider } from "@workos-inc/authkit-react";

export default function Layout() {
    const navigate = useNavigate();
    return (
        <AuthKitProvider
            clientId='client_01J7E5R6JNRYBG12BJQ8J7DGK3'
            apiHostname='rousing-editor-46-staging.authkit.app'
            onRedirectCallback={({ state }) => {
                if (state?.returnTo) {
                    navigate(state.returnTo);
                }
            }}
        >
            <Theme
                accentColor="iris"
                panelBackground="solid"
                style={{ backgroundColor: "var(--gray-1)" }}
            >
        //
            </Theme>
        </AuthKitProvider>
    );
}
