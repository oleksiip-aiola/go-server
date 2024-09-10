import { notifications, NotificationData } from "@mantine/notifications";

export const useToast = () => {
  const showToast = (data: NotificationData) => {
    notifications.show({
      position: "top-right",
      style: { zIndex: 1000 },
      ...data,
    });
  };

  return {
    showToast,
  };
};
