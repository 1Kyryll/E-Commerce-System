import { z } from "zod";

export const addItemSchema = z.object({
  productId: z.string().min(1),
  quantity: z.number().int().positive().max(99),
});
export type AddItemInput = z.infer<typeof addItemSchema>;

export const updateQtySchema = z.object({
  productId: z.string().min(1),
  quantity: z.number().int().min(0).max(99),
});
export type UpdateQtyInput = z.infer<typeof updateQtySchema>;
