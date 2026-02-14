import { useInitialData } from "@/hooks/useInitialData";
import { DataSyncAPI } from "@/lib/data-sync";
import { DataChannel } from "@/store/data-sync.store";

export default function SettingApp() {
  const test = async () => {
    DataSyncAPI.sendToWindow("main", DataChannel.Settings, "settings:update", {
      test: "132",
    });
  };

  const { initialData } = useInitialData<SettingsInitialData>();

  console.log(initialData);

  return (
    <div className="flex items-center justify-center h-screen">
      <h1 className="text-2xl text-black" onClick={test}>
        {initialData?.data.title}
      </h1>
    </div>
  );
}
