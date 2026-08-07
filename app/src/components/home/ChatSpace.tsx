import { useSelector } from "react-redux";
import { selectAuthenticated } from "@/store/auth.ts";
import { infoAuthFooterSelector, infoFooterSelector } from "@/store/info.ts";
import Markdown from "@/components/Markdown.tsx";
import { motion } from "framer-motion";

function Footer() {
  const auth = useSelector(selectAuthenticated);
  const footer = useSelector(infoFooterSelector);
  const auth_footer = useSelector(infoAuthFooterSelector);

  if (auth && auth_footer) {
    // hide footer
    return null;
  }

  return (
    footer.length > 0 && (
      <Markdown
        className={`whitespace-pre-wrap text-secondary text-xs md:text-sm rounded-md bg-background/10 backdrop-blur-sm`}
        acceptHtml={true}
      >
        {footer}
      </Markdown>
    )
  );
}

function ChatSpace() {
  return (
    <motion.div
      className={`chat-product`}
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5 }}
    >
      <motion.div
        className={`space-footer`}
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, delay: 0.8 }}
      >
        <Footer />
      </motion.div>
    </motion.div>
  );
}

export default ChatSpace;
