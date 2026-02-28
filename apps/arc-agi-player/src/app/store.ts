import {
  debugReducer,
  hypercardArtifactsReducer,
  notificationsReducer,
  pluginCardRuntimeReducer,
} from '@hypercard/engine';
import { windowingReducer } from '@hypercard/engine/desktop-core';
import { configureStore } from '@reduxjs/toolkit';
import { arcApi } from '../api/arcApi';
import { arcBridgeReducer, createArcBridgeMiddleware } from '../bridge';
import arcPlayerReducer from '../features/arcPlayer/arcPlayerSlice';

function createArcPlayerStore() {
  const arcBridgeMiddleware = createArcBridgeMiddleware();
  return configureStore({
    reducer: {
      pluginCardRuntime: pluginCardRuntimeReducer,
      arcBridge: arcBridgeReducer,
      windowing: windowingReducer,
      notifications: notificationsReducer,
      debug: debugReducer,
      hypercardArtifacts: hypercardArtifactsReducer,
      arcPlayer: arcPlayerReducer,
      [arcApi.reducerPath]: arcApi.reducer,
    },
    middleware: (getDefaultMiddleware) => getDefaultMiddleware().concat(arcBridgeMiddleware, arcApi.middleware),
  });
}

export const store = createArcPlayerStore();
export { createArcPlayerStore };

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
