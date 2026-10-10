CREATE POLICY "Workspace members can read boards"
ON public.boards
FOR SELECT
TO authenticated
USING (
  EXISTS (
    SELECT 1
    FROM public.workspace_memberships AS wm
    WHERE wm.workspace_id = boards.workspace_id
      AND wm.user_id = (SELECT auth.uid())
  )
);